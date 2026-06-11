# Current State - PamojaBuild Backend

## Last Updated: 2026-06-12

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