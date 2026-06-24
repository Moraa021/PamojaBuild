# Current State - PamojaBuild Backend

## Last Updated: 2026-06-12

## Feature 9: Wallet & Multisig (COMPLETED)

### Implementation Summary

Implemented a 3-of-5 multisig Bitcoin payout system where task funds are cryptographically held on-chain using Bitcoin PSBTs and released via keyholder signatures:

**Files Created/Modified:**
- `backend/db/migrations/005_add_psbt_partial.sql` - Added `psbt_partial` column to `payout_signatures` table.
- `backend/internal/config/config.go` - Added keyholder WIF configuration and Bitcoin RPC host/credential loading.
- `backend/internal/lightning/client.go` - Implemented `PublishTransaction` REST endpoint wrapper for raw transaction broadcasting.
- `backend/internal/bitcoin/client.go` - Created Bitcoin Core RPC client with mock/test fallbacks for UTXO retrieval.
- `backend/internal/models/volunteer.go` & `wallet.go` - Added `PaymentRequest`, `PSBT`, `TxID`, and `PublicKey` fields.
- `backend/internal/wallet/repository.go` - Created database queries for keyholders, signatures, rejections, payout requests, volunteers, and tasks.
- `backend/internal/wallet/service.go` - Implemented 3-of-5 script derivation, base PSBT construction, WIF signature generation, rejection counting/reopening, PSBT finalization/merging, broadcasting, and off-chain volunteer payments.
- `backend/internal/wallet/handler.go` - Implemented `Complete`, `Sign`, and `Reject` handlers.
- `backend/internal/router/router.go` - Registered all HTTP endpoints.

### Wallet & Multisig API

| Method | Route | Auth | Description |
| ----- | ----- | ----- | ----- |
| POST | /tasks/{id}/complete | Required (poster) | Marks task complete, creates payout request, and returns multisig address |
| POST | /wallet/payout/{id}/sign | Required (keyholder) | Signs a payout request (releases funds once 3 signatures are gathered) |
| POST | /wallet/payout/{id}/reject | Required (keyholder) | Rejects a payout request (reopens task back to open status if 3 rejections are gathered) |

### Testing Instructions

Run the unit and integration tests:
```bash
cd backend
go test -v ./internal/wallet/...
```

## Feature 8: Donations (COMPLETED)

### Implementation Summary

Implemented Donations feature to support task donations using the Lightning Network:

**Files Created/Modified:**
- `backend/internal/donations/repository.go` - Database queries (SQLite raw SQL) for donations table.
- `backend/internal/donations/service.go` - Business logic for donating (invoice creation) and calculating task donation totals with status polling.
- `backend/internal/donations/handler.go` - HTTP endpoints for creating a donation and retrieving the total.
- `backend/internal/router/router.go` - Wire up and register route handlers.
- `backend/cmd/server/main.go` - Initialise the HTTP mux with all routes.
- `backend/internal/donations/donations_test.go` - Extensive unit/integration tests with a mock LND TLS REST server.

### Donations API

| Method | Route | Auth | Description |
| ----- | ----- | ----- | ----- |
| POST | /donations/{task_id} | Required | Create Lightning invoice for donation (Accepts `amount_sats` and `is_anonymous` in body) |
| GET | /donations/{task_id}/total | Public | Total confirmed sats for a task (polls pending invoices and calculates sum) |

### Testing Instructions

Run the unit tests:
```bash
cd backend
go test -v ./internal/donations/...
```

## Feature 7: Lightning Integration (COMPLETED)

### Implementation Summary

Implemented thin Lightning client for LND REST API integration:

**Files Created/Modified:**
- `backend/internal/lightning/client.go` - Three function client implementation
- `backend/internal/lightning/client_test.go` - Documentation tests
- `backend/internal/config/config.go` - Config loading with LND settings
- `backend/internal/config/counties.go` - Kenya's 47 counties list
- `backend/internal/models/user.go` - User model with roles
- `backend/internal/models/task.go` - Task model
- `backend/internal/models/donation.go` - Donation model
- `backend/internal/models/volunteer.go` - Volunteer model
- `backend/internal/models/wallet.go` - Payout and Keyholder models
- `backend/internal/db/db.go` - SQLite connection with migrations
- `backend/db/migrations/001_init.sql` - All DB tables
- `backend/internal/middleware/auth.go` - Dummy auth middleware stub
- `backend/cmd/server/main.go` - Main entry point
- `backend/.env.example` - Environment variable template

### Lightning Client API

```go
// CreateInvoice creates a Lightning invoice
CreateInvoice(amountSats int64, memo string) (paymentRequest string, paymentHash string, error)

// PayInvoice pays a Lightning invoice  
PayInvoice(paymentRequest string) error

// CheckPaymentStatus checks if an invoice is settled
CheckPaymentStatus(paymentHash string) (settled bool, error)
```

### Environment Variables Required

```
LND_HOST=localhost:8080
LND_MACAROON_HEX=<hex encoded admin macaroon from Polar>
LND_TLS_CERT_PATH=<path to tls.cert from Polar>
```

### Testing Instructions

1. Install Polar and create a network
2. Get LND macaroon and TLS cert path from Polar
3. Set environment variables in `.env`
4. Run server - functions will be called by donations/wallet services