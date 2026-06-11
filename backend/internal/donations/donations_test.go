package donations_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"PamojaBuild/internal/config"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/donations"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/models"
)

func TestMain(m *testing.M) {
	// Change working directory to backend root so db migrations and config are found
	err := os.Chdir("../../")
	if err != nil {
		fmt.Printf("Warning: failed to change directory to backend root: %v\n", err)
	}
	os.Exit(m.Run())
}

type mockLND struct {
	mu       sync.Mutex
	invoices map[string]bool // payment_hash (base64) -> settled status
}

func newMockLND() *mockLND {
	return &mockLND{
		invoices: make(map[string]bool),
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

	if r.Method == "POST" && r.URL.Path == "/v1/invoices" {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		val, _ := body["value"].(float64)
		amount := int64(val)
		memo, _ := body["memo"].(string)

		payReq := fmt.Sprintf("lnbc%dsatsreq", amount)
		rawHash := []byte(fmt.Sprintf("hash-%d-%s", amount, memo))
		hashBase64 := base64.StdEncoding.EncodeToString(rawHash)

		m.invoices[hashBase64] = false // default not settled

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"payment_request": payReq,
			"r_hash":          hashBase64,
		})
		return
	}

	if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/invoices/") {
		hashBase64 := strings.TrimPrefix(r.URL.Path, "/v1/invoices/")
		settled, exists := m.invoices[hashBase64]
		if !exists {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"settled": settled,
		})
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (m *mockLND) Settle(hashBase64 string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invoices[hashBase64] = true
}

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Seed user and task
	_, err = database.Exec(`
		INSERT INTO users (id, phone, password_hash, role) VALUES (1, '0712345678', 'dummy_hash', 'user');
		INSERT INTO tasks (id, creator_id, title, description, region, max_volunteers, volunteer_mode)
		VALUES (1, 1, 'Test Task', 'This is a test task', 'Nairobi', 5, 'open');
	`)
	if err != nil {
		database.Close()
		t.Fatalf("failed to seed test db: %v", err)
	}

	cleanup := func() {
		database.Close()
	}
	return database, cleanup
}

func setupMockLNDServer(t *testing.T) (*mockLND, *lightning.Client, func()) {
	t.Helper()
	lnd := newMockLND()
	server := httptest.NewTLSServer(lnd)

	// Write server TLS certificate to a temp file
	certBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: server.Certificate().Raw,
	})
	certFile := filepath.Join(t.TempDir(), "tls.cert")
	if err := os.WriteFile(certFile, certBytes, 0600); err != nil {
		server.Close()
		t.Fatalf("failed to write TLS cert to temp file: %v", err)
	}

	// Build lightning client pointing to mock LND
	cfg := &config.Config{
		LNDHost:        strings.TrimPrefix(server.URL, "https://"),
		LNDMacaroonHex: "abc123macaroon",
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

func TestDonate(t *testing.T) {
	database, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, lndClient, lndCleanup := setupMockLNDServer(t)
	defer lndCleanup()

	repo := donations.NewRepository(database.DB)
	service := donations.NewService(repo, lndClient)

	ctx := context.Background()

	// Test 1: Invalid amount
	_, err := service.Donate(ctx, 1, 1, 0, false)
	if err == nil || !strings.Contains(err.Error(), "invalid donation amount") {
		t.Errorf("expected invalid donation amount error, got %v", err)
	}

	// Test 2: Task not found
	_, err = service.Donate(ctx, 999, 1, 1000, false)
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected task not found error, got %v", err)
	}

	// Test 3: Successful donation
	donation, err := service.Donate(ctx, 1, 1, 1000, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if donation.AmountSats != 1000 {
		t.Errorf("expected amount 1000, got %d", donation.AmountSats)
	}
	if donation.Status != string(models.DonationStatusPending) {
		t.Errorf("expected status pending, got %s", donation.Status)
	}
	if donation.PaymentHash == "" || donation.PaymentRequest == "" {
		t.Error("payment request or hash is empty")
	}
}

func TestGetTotalConfirmed(t *testing.T) {
	database, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	lnd, lndClient, lndCleanup := setupMockLNDServer(t)
	defer lndCleanup()

	repo := donations.NewRepository(database.DB)
	service := donations.NewService(repo, lndClient)

	ctx := context.Background()

	// Task 999 not found
	_, err := service.GetTotalConfirmed(ctx, 999)
	if err == nil || !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected task not found error, got %v", err)
	}

	// Empty donations total is 0
	total, err := service.GetTotalConfirmed(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}

	// Create a donation of 500 sats
	_, err = service.Donate(ctx, 1, 1, 500, false)
	if err != nil {
		t.Fatalf("failed to donate: %v", err)
	}

	// Create another donation of 1200 sats
	_, err = service.Donate(ctx, 1, 1, 1200, false)
	if err != nil {
		t.Fatalf("failed to donate: %v", err)
	}

	// Check total: should still be 0 since they are pending
	total, err = service.GetTotalConfirmed(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0 (still pending), got %d", total)
	}

	// Settle the second donation in LND
	// To do this, we need the base64 r_hash that was returned by LND.
	// Since client.go converts base64 to hex, we can convert it back to base64 std encoding.
	// Actually, we can get it from the mock LND's invoices map.
	// Or we can just find which invoice in mockLND has amount 1200.
	lnd.mu.Lock()
	var d2HashBase64 string
	for k := range lnd.invoices {
		decoded, _ := base64.StdEncoding.DecodeString(k)
		if strings.Contains(string(decoded), "1200") {
			d2HashBase64 = k
			break
		}
	}
	lnd.mu.Unlock()

	if d2HashBase64 == "" {
		t.Fatal("could not find mock invoice for d2")
	}

	lnd.Settle(d2HashBase64)

	// Now check total: should be 1200
	total, err = service.GetTotalConfirmed(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1200 {
		t.Errorf("expected total 1200, got %d", total)
	}

	// Settle the first donation (500 sats)
	lnd.mu.Lock()
	var d1HashBase64 string
	for k := range lnd.invoices {
		decoded, _ := base64.StdEncoding.DecodeString(k)
		if strings.Contains(string(decoded), "500") {
			d1HashBase64 = k
			break
		}
	}
	lnd.mu.Unlock()

	lnd.Settle(d1HashBase64)

	// Now check total: should be 1700
	total, err = service.GetTotalConfirmed(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1700 {
		t.Errorf("expected total 1700, got %d", total)
	}
}
