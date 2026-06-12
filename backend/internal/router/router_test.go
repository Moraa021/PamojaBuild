package router_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"PamojaBuild/internal/config"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/models"
	"PamojaBuild/internal/router"
)

func TestMain(m *testing.M) {
	// Change working directory to backend root so migrations are found
	err := os.Chdir("../../")
	if err != nil {
		fmt.Printf("Warning: failed to change directory to backend root: %v\n", err)
	}
	os.Exit(m.Run())
}

type mockLND struct {
	mu           sync.Mutex
	invoices     map[string]bool // payment_hash (base64) -> settled status
	payments     []string
	publishedTxs []string
}

func newMockLND() *mockLND {
	return &mockLND{
		invoices:     make(map[string]bool),
		payments:     make([]string, 0),
		publishedTxs: make([]string, 0),
	}
}

func (m *mockLND) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	macaroon := r.Header.Get("Grpc-Metadata-Macaroon")
	if macaroon == "" {
		http.Error(w, "missing macaroon", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Create Invoice
	if r.Method == "POST" && r.URL.Path == "/v1/invoices" {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		
		memo := ""
		if m, ok := req["memo"].(string); ok {
			memo = m
		}
		
		// Generate dummy payment hash and request
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", memo, time.Now().UnixNano())))
		hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
		
		m.invoices[hashBase64] = true // Auto-settle in mock
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"payment_request": "lnbc_test_payment_request_" + hashBase64[:10],
			"r_hash":          hashBase64,
		})
		return
	}

	// Pay Invoice
	if r.Method == "POST" && r.URL.Path == "/v1/pay" {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		pr := req["payment_request"].(string)
		m.payments = append(m.payments, pr)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txid":"dummy_pay_txid"}`))
		return
	}

	// Check Payment Status
	if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/invoices/") {
		parts := strings.Split(r.URL.Path, "/")
		hashBase64 := parts[len(parts)-1]
		
		settled := m.invoices[hashBase64]
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"settled": settled,
		})
		return
	}

	// Publish transaction
	if r.Method == "POST" && r.URL.Path == "/v2/wallet/tx" {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		txHex := req["tx_hex"].(string)
		m.publishedTxs = append(m.publishedTxs, txHex)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txid":"e2e_broadcasted_txid"}`))
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "e2e_test.db")
	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	cleanup := func() {
		database.Close()
	}
	return database, cleanup
}

func setupMockLND(t *testing.T) (*mockLND, *lightning.Client, func()) {
	t.Helper()
	lnd := newMockLND()
	server := httptest.NewTLSServer(lnd)

	// Write TLS cert
	certBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: server.Certificate().Raw,
	})
	certFile := filepath.Join(t.TempDir(), "tls.cert")
	if err := os.WriteFile(certFile, certBytes, 0600); err != nil {
		server.Close()
		t.Fatalf("failed to write TLS cert: %v", err)
	}

	cfg := &config.Config{
		LNDHost:        strings.TrimPrefix(server.URL, "https://"),
		LNDMacaroonHex: "e2emacaroon",
		LNDTLSCertPath: certFile,
	}

	client, err := lightning.NewClient(cfg)
	if err != nil {
		server.Close()
		t.Fatalf("failed to create lightning client: %v", err)
	}

	cleanup := func() {
		server.Close()
	}
	return lnd, client, cleanup
}

func registerAndLogin(t *testing.T, serverURL, phone, password string) string {
	t.Helper()
	// 1. Register
	regBody, _ := json.Marshal(map[string]string{
		"phone":    phone,
		"password": password,
	})
	resp, err := http.Post(serverURL+"/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected registration status %d: %s", resp.StatusCode, string(b))
	}

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"phone":    phone,
		"password": password,
	})
	resp, err = http.Post(serverURL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected login status %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	
	// Unpack standardized success response {"data": {"token": "...", "user": ...}}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("login response missing data field: %+v", result)
	}
	token, ok := data["token"].(string)
	if !ok {
		t.Fatalf("login data missing token field: %+v", data)
	}
	return token
}

func TestEndToEndIntegrationFlow(t *testing.T) {
	testDB, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, lndClient, lndCleanup := setupMockLND(t)
	defer lndCleanup()

	// Seed keyholders in SQLite DB with a valid password hash of "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}
	passHash := string(hashedPassword)

	keyholderPhones := []string{
		"0711000001",
		"0711000002",
		"0711000003",
		"0711000004",
		"0711000005",
	}

	for i, phone := range keyholderPhones {
		// Insert user
		uid := int64(i + 2) // creator has ID 1
		_, err := testDB.Exec(
			"INSERT INTO users (id, phone, password_hash, role) VALUES (?, ?, ?, 'keyholder')",
			uid, phone, passHash,
		)
		if err != nil {
			t.Fatalf("failed to seed keyholder user: %v", err)
		}
	}

	// Setup Router
	jwtSecret := "e2e-jwt-secret-key-1234567890"
	r := router.SetupRouter(jwtSecret, testDB.DB, lndClient)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Register and Login Task Creator
	creatorToken := registerAndLogin(t, ts.URL, "0712345678", "password123")

	// 2. Create Task (requires auth)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", "Clean Kibuye Market")
	_ = writer.WriteField("description", "Join us to clean up the overgrown market square")
	_ = writer.WriteField("region", "Kisumu")
	_ = writer.WriteField("category", "cleanup")
	_ = writer.WriteField("max_volunteers", "2") // small cap to trigger status change easily
	_ = writer.WriteField("volunteer_mode", "open")
	_ = writer.WriteField("goal_sats", "1000")
	writer.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/tasks", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+creatorToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created on task creation, got %d: %s", resp.StatusCode, string(b))
	}

	var taskRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&taskRes)
	taskData, ok := taskRes["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected task response payload: %+v", taskRes)
	}
	taskID := int64(taskData["id"].(float64))
	if taskData["status"].(string) != "open" {
		t.Errorf("expected task status to be 'open', got %s", taskData["status"].(string))
	}

	// 3. Make a donation of 1000 sats
	donReqBody, _ := json.Marshal(map[string]interface{}{
		"amount_sats":  1000,
		"is_anonymous": false,
	})
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/donations/%d", ts.URL, taskID), bytes.NewReader(donReqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorToken)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to create donation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created on donation, got %d: %s", resp.StatusCode, string(b))
	}

	// Check donation total (polls and confirms)
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/donations/%d/total", ts.URL, taskID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to get total donations: %v", err)
	}
	defer resp.Body.Close()

	var totalRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&totalRes)
	totalData := totalRes["data"].(map[string]interface{})
	if totalData["total_sats"].(float64) != 1000 {
		t.Errorf("expected 1000 sats donation, got %f", totalData["total_sats"].(float64))
	}

	// 4. Volunteer 1 Applies (Auto-approved because mode is open)
	vol1Token := registerAndLogin(t, ts.URL, "0799999991", "password123")
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/tasks/%d/volunteer", ts.URL, taskID), nil)
	req.Header.Set("Authorization", "Bearer "+vol1Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("volunteer 1 failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created on volunteer 1, got %d: %s", resp.StatusCode, string(b))
	}

	// Verify task is still open (1/2 volunteers)
	var task models.Task
	err = testDB.QueryRow("SELECT status, max_volunteers FROM tasks WHERE id = ?", taskID).Scan(&task.Status, &task.MaxVolunteers)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if task.Status != "open" {
		t.Errorf("expected task status to be 'open' after 1 volunteer, got %s", task.Status)
	}

	// 5. Volunteer 2 Applies (Auto-approved, reaches cap, transitions to in_progress)
	vol2Token := registerAndLogin(t, ts.URL, "0799999992", "password123")
	
	// Add invoice for volunteer 1 and 2 to get paid later
	_, _ = testDB.Exec("UPDATE volunteers SET payment_request = 'lnbc100v1...' WHERE user_id = (SELECT id FROM users WHERE phone = '0799999991')")
	
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/tasks/%d/volunteer", ts.URL, taskID), nil)
	req.Header.Set("Authorization", "Bearer "+vol2Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("volunteer 2 failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created on volunteer 2, got %d: %s", resp.StatusCode, string(b))
	}

	_, _ = testDB.Exec("UPDATE volunteers SET payment_request = 'lnbc100v2...' WHERE user_id = (SELECT id FROM users WHERE phone = '0799999992')")

	// Verify task status has automatically transitioned to 'in_progress'!
	err = testDB.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&task.Status)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if task.Status != "in_progress" {
		t.Errorf("expected task status to transition to 'in_progress' after cap met, got %s", task.Status)
	}

	// 6. Creator completes task (Triggers payout request, status transitions to pending_verification)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/tasks/%d/complete", ts.URL, taskID), nil)
	req.Header.Set("Authorization", "Bearer "+creatorToken)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("complete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201 Created on complete, got %d: %s", resp.StatusCode, string(b))
	}

	var payoutRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&payoutRes)
	payoutData := payoutRes["data"].(map[string]interface{})
	payoutReqID := int64(payoutData["payout_request_id"].(float64))

	// Verify task status is 'pending_verification'
	err = testDB.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&task.Status)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if task.Status != "pending_verification" {
		t.Errorf("expected task status to be 'pending_verification', got %s", task.Status)
	}

	// 7. Keyholder 1 signs (1/3 signatures)
	kh1Token := registerAndLogin(t, ts.URL, "0711000001", "password123")
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/wallet/payout/%d/sign", ts.URL, payoutReqID), nil)
	req.Header.Set("Authorization", "Bearer "+kh1Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("kh1 sign failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK on kh1 sign, got %d: %s", resp.StatusCode, string(b))
	}

	// 8. Keyholder 2 signs (2/3 signatures)
	kh2Token := registerAndLogin(t, ts.URL, "0711000002", "password123")
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/wallet/payout/%d/sign", ts.URL, payoutReqID), nil)
	req.Header.Set("Authorization", "Bearer "+kh2Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("kh2 sign failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK on kh2 sign, got %d: %s", resp.StatusCode, string(b))
	}

	// Verify payout status is still pending (2/3 approvals)
	var payoutStatus string
	err = testDB.QueryRow("SELECT status FROM payout_requests WHERE id = ?", payoutReqID).Scan(&payoutStatus)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if payoutStatus != "pending" {
		t.Errorf("expected payout request status to still be 'pending', got %s", payoutStatus)
	}

	// 9. Keyholder 3 signs (3/3 signatures, triggers broadcast, task completed, volunteers paid)
	kh3Token := registerAndLogin(t, ts.URL, "0711000003", "password123")
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/wallet/payout/%d/sign", ts.URL, payoutReqID), nil)
	req.Header.Set("Authorization", "Bearer "+kh3Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("kh3 sign failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK on kh3 sign, got %d: %s", resp.StatusCode, string(b))
	}

	// 10. Verify post-threshold state
	// Verify payout status is 'released'
	err = testDB.QueryRow("SELECT status FROM payout_requests WHERE id = ?", payoutReqID).Scan(&payoutStatus)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if payoutStatus != "released" {
		t.Errorf("expected payout request status to be 'released', got %s", payoutStatus)
	}

	// Verify task status is 'completed'
	err = testDB.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&task.Status)
	if err != nil {
		t.Fatalf("db error: %v", err)
	}
	if task.Status != "completed" {
		t.Errorf("expected task status to transition to 'completed', got %s", task.Status)
	}

	// Verify volunteers statuses are 'paid'
	var vol1Status, vol2Status string
	_ = testDB.QueryRow("SELECT status FROM volunteers WHERE user_id = (SELECT id FROM users WHERE phone = '0799999991')").Scan(&vol1Status)
	_ = testDB.QueryRow("SELECT status FROM volunteers WHERE user_id = (SELECT id FROM users WHERE phone = '0799999992')").Scan(&vol2Status)

	if vol1Status != "paid" {
		t.Errorf("expected volunteer 1 status to transition to 'paid', got %s", vol1Status)
	}
	if vol2Status != "paid" {
		t.Errorf("expected volunteer 2 status to transition to 'paid', got %s", vol2Status)
	}
}
