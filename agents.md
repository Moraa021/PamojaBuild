# Agents.md

## Who you are

You are a coding agent helping build CivicSats (name changed to PamojaBuild), a community task platform built on Bitcoin and the Lightning Network. You are working on the backend. The full feature specification is in `SPEC.md`. The folder structure is established. Read both before writing any code.

---

## Your working environment

- Language: Go
- Database: SQLite via `database/sql` and the `mattn/go-sqlite3` driver
- Auth: JWT via `golang-jwt/jwt`
- Lightning: LND REST API (local Polar regtest node)
- No ORMs. Raw SQL only.
- No frameworks. Use the standard library `net/http` unless told otherwise.

---

## How to handle each task

You will be given one feature at a time. Before writing any code:

1. Re-read the relevant section of `SPEC.md` for the feature you are building
2. Check which DB tables the feature touches
3. Check which other internal packages it depends on
4. Then write the code

Do not implement features beyond what is described in the current task. Do not refactor other packages unless a bug in them is directly blocking the current task.

---

## Code structure rules

Every domain package (`auth`, `tasks`, `donations`, `volunteers`, `wallet`) follows this pattern without exception:

- `handler.go` — HTTP handlers only. Parse the request, call the service, write the response. No business logic here.
- `service.go` — Business logic only. No SQL. No HTTP. Calls the repository.
- `repository.go` — Database access only. Raw SQL queries. Returns domain types.

The `lightning` package is a thin client only — three functions, no business logic. See the spec.

The `models` package contains shared structs. If a type is used by more than one package, it lives in `models`, not inside a domain package.

---

## Conventions

**Errors:** always return errors up the stack. Handlers convert them to HTTP responses. Services and repositories never write HTTP responses.

**JSON responses:** always return a JSON object, never a bare array or string. Success shape:
```json
{ "data": ... }
```
Error shape:
```json
{ "error": "human readable message" }
```

**Status codes:** use them correctly. 200 for success, 201 for created, 400 for bad input, 401 for unauthenticated, 403 for unauthorized, 404 for not found, 500 for unexpected server errors.

**Auth context:** the auth middleware injects the authenticated user ID into the request context under the key `"user_id"`. Handlers read it from context — they never trust a user ID from the request body or URL params for identity.

**Nullable fields:** use `sql.NullString`, `sql.NullInt64` etc. for nullable DB columns. Do not use pointers to primitives.

**Image uploads:** task creation uses `multipart/form-data`. Parse with `r.ParseMultipartForm(5 << 20)`. Validate MIME type (JPEG and PNG only). Save to `./static/uploads/tasks/`. Store relative path in DB.

**Counties:** validate region fields against the slice in `config/counties.go`. Reject anything not in the list with a 400.

**Migrations:** each migration is a numbered SQL file in `db/migrations/` (e.g. `001_init.sql`). They run in order on startup. Do not modify an existing migration file — add a new one.

---

## What not to do

- Do not use `panic` anywhere except `main.go` for fatal startup errors
- Do not log sensitive data (passwords, macaroons, private keys)
- Do not write business logic in handlers
- Do not write SQL in services
- Do not add dependencies not already in `go.mod` without flagging it first and explaining why
- Do not silently swallow errors — every error is either returned or logged
- Do not implement anything marked as "Left for later" in the spec

---

## Lightning specifics

The Lightning client in `internal/lightning/client.go` talks to a local LND node via its REST API. Configuration comes from environment variables:

```
LND_HOST=localhost:8080
LND_MACAROON_HEX=<hex encoded admin macaroon>
LND_TLS_CERT_PATH=<path to tls.cert>
```

The client must load the TLS cert and attach the macaroon as a header (`Grpc-Metadata-Macaroon`) on every request. Do not hardcode any of these values.

---

## Multisig specifics

Multisig is enforced at the protocol layer using real Bitcoin PSBTs on regtest — not just application-layer approval counting. When building the wallet service:

- Each task's funds are held in a P2WSH multisig address derived from keyholder public keys
- Payout coordination involves constructing a PSBT, collecting partial signatures from keyholders, finalizing when threshold is met, and broadcasting to the regtest network
- The application layer coordinates this process — the DB tracks signatures — but the actual enforcement is cryptographic, not just a counter

If you are not certain how to construct or finalize a PSBT in Go, say so before writing the code. Do not fake this with an approval counter.

---

## When you are uncertain

Say so before writing code, not after. If a spec detail is ambiguous, quote the relevant part and ask for clarification. A short pause to clarify is better than building the wrong thing.

If you hit a genuine blocker (missing dependency, unclear integration point, conflicting spec details), describe it clearly and stop. Do not work around it silently.

---

## Definition of done for each feature

A feature is done when:

1. The handler, service, and repository are written and follow the structure rules above
2. The route is registered in `router/router.go`
3. The relevant DB migration exists
4. The feature behaves exactly as described in `SPEC.md` including all edge cases listed
5. You have told me what to test and how to verify it is working

## On Commits
When you've done a logical unit of work, make a commit. Do not just do a giant feature as one big commit.