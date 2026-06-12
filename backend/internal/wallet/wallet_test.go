package wallet_test

import (
	"context"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/btcsuite/btcd/btcutil/psbt"

	"PamojaBuild/internal/config"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/models"
	"PamojaBuild/internal/wallet"
)

func TestMain(m *testing.M) {
	// Change working directory to backend root so migrations and config are found
	err := os.Chdir("../../")
	if err != nil {
		fmt.Printf("Warning: failed to change directory to backend root: %v\n", err)
	}
	os.Exit(m.Run())
}

type mockLND struct {
	mu            sync.Mutex
	payments      []string // payment requests paid
	publishedTxs  []string // tx hexes published
}

func newMockLND() *mockLND {
	return &mockLND{
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

	if r.Method == "POST" && r.URL.Path == "/v1/pay" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txid":"dummy_pay_txid"}`))
		return
	}

	if r.Method == "POST" && r.URL.Path == "/v2/wallet/tx" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"txid":"dummy_pub_txid"}`))
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	// Seed users, task, volunteers, and donations
	_, err = database.Exec(`
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (1, '0712345678', 'dummy_hash', 'user');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (2, '0722222222', 'dummy_hash', 'keyholder');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (3, '0733333333', 'dummy_hash', 'keyholder');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (4, '0744444444', 'dummy_hash', 'keyholder');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (5, '0755555555', 'dummy_hash', 'keyholder');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (6, '0766666666', 'dummy_hash', 'keyholder');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (11, '0711111111', 'dummy_hash', 'user');
		INSERT OR IGNORE INTO users (id, phone, password_hash, role) VALUES (12, '0712121212', 'dummy_hash', 'user');

		INSERT OR IGNORE INTO tasks (id, creator_id, title, description, region, max_volunteers, volunteer_mode, status)
		VALUES (1, 1, 'Test Task', 'This is a test task', 'Nairobi', 5, 'open', 'in_progress');

		INSERT OR IGNORE INTO volunteers (id, task_id, user_id, status, payment_request)
		VALUES (1, 1, 11, 'approved', 'lnbc100satsvolunteer1');
		INSERT OR IGNORE INTO volunteers (id, task_id, user_id, status, payment_request)
		VALUES (2, 1, 12, 'approved', 'lnbc100satsvolunteer2');

		INSERT OR IGNORE INTO donations (id, task_id, donor_id, amount_sats, payment_hash, payment_request, status)
		VALUES (1, 1, 1, 1000, 'hash1', 'req1', 'confirmed');
		INSERT OR IGNORE INTO donations (id, task_id, donor_id, amount_sats, payment_hash, payment_request, status)
		VALUES (2, 1, 1, 500, 'hash2', 'req2', 'confirmed');
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

	// Set env vars so wallet service can broadcast via broadcastTx
	os.Setenv("LND_HOST", cfg.LNDHost)
	os.Setenv("LND_MACAROON_HEX", cfg.LNDMacaroonHex)
	os.Setenv("LND_TLS_CERT_PATH", cfg.LNDTLSCertPath)

	client, err := lightning.NewClient(cfg)
	if err != nil {
		server.Close()
		t.Fatalf("failed to create lightning client: %v", err)
	}

	cleanup := func() {
		server.Close()
		os.Unsetenv("LND_HOST")
		os.Unsetenv("LND_MACAROON_HEX")
		os.Unsetenv("LND_TLS_CERT_PATH")
	}
	return lnd, client, cleanup
}

func TestWalletLifecycle(t *testing.T) {
	database, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, lndClient, lndCleanup := setupMockLNDServer(t)
	defer lndCleanup()

	repo := wallet.NewRepository(database.DB)
	service := wallet.NewService(repo, lndClient)

	ctx := context.Background()

	// 1. Verify Keyholders Initialization
	keyholders, err := repo.GetKeyholders(ctx)
	if err != nil {
		t.Fatalf("failed to get keyholders: %v", err)
	}
	if len(keyholders) != 5 {
		t.Fatalf("expected 5 keyholders, got %d", len(keyholders))
	}
	for _, kh := range keyholders {
		if kh.PublicKey == "" {
			t.Errorf("keyholder %d has empty public key", kh.ID)
		}
	}

	// 2. Test CompleteTask (Create Payout Request)
	// User 99 (not task poster) should fail
	_, err = service.CompleteTask(ctx, 1, 99)
	if err == nil || !strings.Contains(err.Error(), "user is not the creator") {
		t.Errorf("expected creator permission error, got %v", err)
	}

	// Creator (User 1) completes task
	pr, err := service.CompleteTask(ctx, 1, 1)
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	if pr.TaskID != 1 {
		t.Errorf("expected task id 1, got %d", pr.TaskID)
	}
	if pr.TotalSats != 1500 {
		t.Errorf("expected total sats 1500, got %d", pr.TotalSats)
	}
	if pr.Status != string(models.PayoutStatusPending) {
		t.Errorf("expected pending status, got %s", pr.Status)
	}
	if pr.PSBT == "" {
		t.Fatal("expected PSBT to be generated, got empty string")
	}

	// Verify PSBT structure
	p, err := psbt.NewFromRawBytes(strings.NewReader(pr.PSBT), true)
	if err != nil {
		t.Fatalf("failed to parse generated PSBT: %v", err)
	}
	if len(p.UnsignedTx.TxIn) != 1 {
		t.Errorf("expected 1 input, got %d", len(p.UnsignedTx.TxIn))
	}
	if len(p.UnsignedTx.TxOut) != 2 {
		t.Errorf("expected 2 outputs (for 2 volunteers), got %d", len(p.UnsignedTx.TxOut))
	}

	// Check output values (1500 / 2 = 750 each)
	for _, out := range p.UnsignedTx.TxOut {
		if out.Value != 750 {
			t.Errorf("expected volunteer share to be 750, got %d", out.Value)
		}
	}

	// Verify task status updated to pending_verification
	_, taskStatus, err := repo.GetTaskCreatorAndStatus(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get task status: %v", err)
	}
	if taskStatus != "pending_verification" {
		t.Errorf("expected task status pending_verification, got %s", taskStatus)
	}

	// 3. Test Keyholder Signing
	// Non-keyholder (User 1) cannot sign
	_, err = service.SignPayoutRequest(ctx, pr.ID, 1)
	if err == nil || !strings.Contains(err.Error(), "user is not a keyholder") {
		t.Errorf("expected non-keyholder error, got %v", err)
	}

	// Sign 1: Keyholder 1 (User 2)
	pr, err = service.SignPayoutRequest(ctx, pr.ID, 2)
	if err != nil {
		t.Fatalf("keyholder 1 failed to sign: %v", err)
	}
	if pr.Status != string(models.PayoutStatusPending) {
		t.Errorf("expected payout request to remain pending, got %s", pr.Status)
	}

	// Try to sign again (should fail)
	_, err = service.SignPayoutRequest(ctx, pr.ID, 2)
	if err == nil || !strings.Contains(err.Error(), "already signed") {
		t.Errorf("expected already signed error, got %v", err)
	}

	// Sign 2: Keyholder 2 (User 3)
	pr, err = service.SignPayoutRequest(ctx, pr.ID, 3)
	if err != nil {
		t.Fatalf("keyholder 2 failed to sign: %v", err)
	}
	if pr.Status != string(models.PayoutStatusPending) {
		t.Errorf("expected payout request to remain pending, got %s", pr.Status)
	}

	// Sign 3: Keyholder 3 (User 4) -> This meets the threshold of 3 approvals!
	pr, err = service.SignPayoutRequest(ctx, pr.ID, 4)
	if err != nil {
		t.Fatalf("keyholder 3 failed to sign: %v", err)
	}

	// Check released status
	if pr.Status != string(models.PayoutStatusReleased) {
		t.Errorf("expected payout request status to be released, got %s", pr.Status)
	}
	if pr.TxID == "" {
		t.Error("expected transaction ID to be populated")
	}

	// Verify task status updated to completed
	_, taskStatus, err = repo.GetTaskCreatorAndStatus(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get task status: %v", err)
	}
	if taskStatus != "completed" {
		t.Errorf("expected task status completed, got %s", taskStatus)
	}

	// Verify volunteers statuses updated to paid
	vols, err := repo.GetApprovedVolunteers(ctx, 1)
	if err != nil {
		t.Fatalf("failed to query volunteers: %v", err)
	}
	// Note: since the payout has been released, GetApprovedVolunteers (which only selects status='approved')
	// should return 0 volunteers because their status has transitioned to 'paid'!
	if len(vols) != 0 {
		t.Errorf("expected 0 approved volunteers remaining (since they should be paid), got %d", len(vols))
	}

	// Query directly from database to verify status is 'paid'
	var statuses []string
	rows, err := database.QueryContext(ctx, "SELECT status FROM volunteers WHERE task_id = 1")
	if err != nil {
		t.Fatalf("failed to query volunteers table directly: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		rows.Scan(&s)
		statuses = append(statuses, s)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 volunteers, got %d", len(statuses))
	}
	for _, s := range statuses {
		if s != string(models.VolunteerStatusPaid) {
			t.Errorf("expected volunteer status paid, got %s", s)
		}
	}
}

func TestWalletRejection(t *testing.T) {
	database, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, lndClient, lndCleanup := setupMockLNDServer(t)
	defer lndCleanup()

	repo := wallet.NewRepository(database.DB)
	service := wallet.NewService(repo, lndClient)

	ctx := context.Background()

	// Poster (User 1) completes task
	pr, err := service.CompleteTask(ctx, 1, 1)
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	// Reject 1: Keyholder 1 (User 2)
	pr, err = service.RejectPayoutRequest(ctx, pr.ID, 2)
	if err != nil {
		t.Fatalf("keyholder 1 failed to reject: %v", err)
	}
	if pr.Status != string(models.PayoutStatusPending) {
		t.Errorf("expected pending, got %s", pr.Status)
	}

	// Reject 2: Keyholder 2 (User 3)
	pr, err = service.RejectPayoutRequest(ctx, pr.ID, 3)
	if err != nil {
		t.Fatalf("keyholder 2 failed to reject: %v", err)
	}

	// Reject 3: Keyholder 3 (User 4) -> Threshold met!
	pr, err = service.RejectPayoutRequest(ctx, pr.ID, 4)
	if err != nil {
		t.Fatalf("keyholder 3 failed to reject: %v", err)
	}

	if pr.Status != string(models.PayoutStatusRejected) {
		t.Errorf("expected rejected status, got %s", pr.Status)
	}

	// Verify task status re-opened to open
	_, taskStatus, err := repo.GetTaskCreatorAndStatus(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if taskStatus != "open" {
		t.Errorf("expected task status open, got %s", taskStatus)
	}
}

func TestGhostKeyholdersAutoRelease(t *testing.T) {
	database, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, lndClient, lndCleanup := setupMockLNDServer(t)
	defer lndCleanup()

	repo := wallet.NewRepository(database.DB)
	service := wallet.NewService(repo, lndClient)

	ctx := context.Background()

	// Poster completes task
	pr, err := service.CompleteTask(ctx, 1, 1)
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	// Sign 1: Keyholder 1 (User 2)
	pr, err = service.SignPayoutRequest(ctx, pr.ID, 2)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	// Sign 2: Keyholder 2 (User 3)
	pr, err = service.SignPayoutRequest(ctx, pr.ID, 3)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	// Now we have 2 signatures (less than 3).
	// Let's backdate the payout request to be 4 days old (older than 72 hours).
	_, err = database.Exec(`
		UPDATE payout_requests 
		SET created_at = datetime('now', '-4 days') 
		WHERE id = ?
	`, pr.ID)
	if err != nil {
		t.Fatalf("failed to backdate payout request: %v", err)
	}

	// Trigger CheckGhostKeyholders manually to verify
	service.CheckGhostKeyholders(ctx)

	// Check if status updated to released
	prUpdated, err := repo.GetPayoutRequest(ctx, pr.ID)
	if err != nil {
		t.Fatalf("failed to query payout request: %v", err)
	}
	if prUpdated.Status != string(models.PayoutStatusReleased) {
		t.Errorf("expected released, got %s", prUpdated.Status)
	}
	if prUpdated.TxID != "auto-released" {
		t.Errorf("expected tx_id to be 'auto-released', got %s", prUpdated.TxID)
	}

	// Verify task completed
	_, taskStatus, err := repo.GetTaskCreatorAndStatus(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if taskStatus != "completed" {
		t.Errorf("expected task status completed, got %s", taskStatus)
	}
}
