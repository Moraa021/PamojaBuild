# **PamojaBuild — Integration Notes & Project Inventory**

This document provides a detailed inventory of the PamojaBuild codebase, mapping what has been built against the feature specification, highlighting gaps and variations in the frontend and backend, and outlining the next steps required to achieve full integration.

---

## **1. Feature Specification Inventory**

Below is a detailed audit of each feature block from [SPEC.md.md](file:///home/frank/projects/repos/PamojaBuild/SPEC.md.md) showing its implementation status in the backend.

| Feature Domain | Spec Requirement | Backend Status | Code Reference |
| :--- | :--- | :--- | :--- |
| **Auth & Users** | Register with phone + password | **Implemented** | [auth/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/auth/service.go#L31) |
| | Login returns a JWT | **Implemented** | [auth/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/auth/service.go#L62) |
| | JWT Validation Middleware | **Implemented** | [middleware/auth.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/middleware/auth.go#L25) |
| | Seed Keyholders on Startup | **Partially Implemented** (Has placeholder hash; see Gaps) | [server/main.go](file:///home/frank/projects/repos/PamojaBuild/backend/cmd/server/main.go#L55) |
| **Tasks** | Create task (multipart/form-data with image) | **Implemented** | [tasks/handler.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/handler.go#L23) |
| | Serve uploaded task images statically | **Implemented** | [router/router.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/router/router.go#L90) |
| | List tasks with filters (region, status, category) | **Implemented** | [tasks/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/repository.go#L23) |
| | Single task view | **Implemented** | [tasks/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/repository.go#L58) |
| | Status lifecycle (`open → in_progress → pending_verification → completed`) | **Broken** (No code transitions tasks to `in_progress`; see Gaps) | [tasks/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/repository.go#L64) |
| | County validation | **Implemented** | [tasks/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/service.go#L37) |
| | Volunteer cap adjustment (Raise only) | **Implemented** | [tasks/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/tasks/service.go#L81) |
| **Donations** | Create Lightning Invoice for Donation | **Implemented** | [donations/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/donations/service.go#L35) |
| | Poll and confirm invoice payment | **Implemented** | [donations/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/donations/service.go#L74) |
| | Get total confirmed sats for task | **Implemented** | [donations/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/donations/repository.go#L102) |
| | Public/Anonymous donations option | **Implemented** | [donations/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/donations/repository.go#L38) |
| **Volunteers** | Apply to volunteer for task | **Implemented** | [volunteers/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/volunteers/service.go#L19) |
| | Open Mode (Auto-approved up to cap) | **Implemented** | [volunteers/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/volunteers/service.go#L45) |
| | Approval Mode (Poster approves manually) | **Implemented** | [volunteers/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/volunteers/service.go#L55) |
| | volunteer status (`pending → approved → paid`) | **Implemented** | [volunteers/repository.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/volunteers/repository.go#L44) |
| **Wallet** | Complete Task (creates payout request & derives multisig) | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L48) |
| | Keyholder sign/reject payout | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L205) |
| | 3-of-5 script derivation, PSBT creation | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L99) |
| | PSBT partial signing + finalization & broadcast | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L386) |
| | Release payout equally to volunteers via Lightning | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L494) |
| | Ghost keyholder timeout (72h, auto-release with 2+ sigs) | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L508) |
| | Rejections handling (3 rejections reopens task) | **Implemented** | [wallet/service.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/wallet/service.go#L370) |
| **Lightning** | LND REST thin client (Create, Pay, Check, Broadcast) | **Implemented** | [lightning/client.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/lightning/client.go) |
| **Location** | Hardcoded 47 counties config | **Implemented** | [config/counties.go](file:///home/frank/projects/repos/PamojaBuild/backend/internal/config/counties.go) |

---

## **2. Key Backend Implementation Gaps & Variations**

During the codebase audit, several critical bugs, gaps, and design variations were identified in the backend:

### **A. Task Status Transition to `in_progress` is Missing**
* **Issue**: A task is created with status `'open'`. The wallet service completion endpoint `POST /tasks/{id}/complete` enforces that the task status must be `'in_progress'`. However, there is no code in the `volunteers` or `tasks` service that changes the task status to `'in_progress'` when the volunteer cap is reached.
* **Impact**: In the current state, a task remains `'open'` forever. A user can never successfully complete a task via the API because it will fail validation.
* **Resolution**: Modify `volunteers.Service.Apply` and `volunteers.Service.Approve` to check if the count of approved volunteers has reached `max_volunteers`. If it has, use the tasks repository to update the task status to `'in_progress'`.

### **B. API Response Shape Inconsistency**
* **Issue**: The [agents.md](file:///home/frank/projects/repos/PamojaBuild/agents.md) file states that all successful JSON responses must wrap their payload in a `"data"` key: `{"data": ...}`, and error responses must return `{"error": "..."}`.
  * The `donations` and `wallet` packages follow this rule perfectly.
  * The `auth`, `tasks`, and `volunteers` packages return raw JSON structs (e.g. `models.Task` directly) without the `"data"` key wrapper.
  * Furthermore, `tasks`, `volunteers`, and `auth` packages frequently write plain text error messages using `http.Error` instead of returning a JSON error response.
* **Impact**: The frontend has to handle two completely different response styles depending on which endpoint it hits, leading to brittle integration code.
* **Resolution**: Update the handlers in `auth`, `tasks`, and `volunteers` to wrap successful responses in `{"data": ...}` and error responses in `{"error": ...}`.

### **C. Seeded Keyholder Authentication Block**
* **Issue**: The server automatically seeds 5 keyholders on startup. However, their password hashes are set to the raw string `"$2a$10$tempHashForSeededKeyholders"`. Since this is a dummy string and not a valid bcrypt hash, it is impossible to log in as any of the seeded keyholders.
* **Impact**: Keyholders cannot log in through `/auth/login` to sign or reject payout requests.
* **Resolution**: Update `seedKeyholders()` in [server/main.go](file:///home/frank/projects/repos/PamojaBuild/backend/cmd/server/main.go#L55) to use a valid bcrypt hash of a known default password (e.g., `"password123"` or `"keyholderpass"`).

### **D. Missing CORS Support**
* **Issue**: The backend uses the standard Go `net/http` server without setting any CORS (Cross-Origin Resource Sharing) headers.
* **Impact**: Browsers will block any API requests from the frontend to the backend if the frontend is opened directly as a file (`file:///...`) or served from a different port/host.
* **Resolution**: Add a global CORS middleware in `router/router.go` to inject headers (`Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`) and handle preflight `OPTIONS` requests.

---

## **3. Frontend State & Variations**

### **Current State**
* **Landing Page**: [index.html](file:///home/frank/projects/repos/PamojaBuild/frontend/index.html) is fully implemented. It features a premium, responsive design styled with the CSS tokens and styles defined in [css/styles.css](file:///home/frank/projects/repos/PamojaBuild/frontend/css/styles.css).
* **Missing Interactive Pages**: The core application pages are completely blank:
  * [tasks.html](file:///home/frank/projects/repos/PamojaBuild/frontend/tasks.html) (0 bytes) — intended for listing, filtering, and searching tasks.
  * [app.html](file:///home/frank/projects/repos/PamojaBuild/frontend/app.html) (0 bytes) — the core app interface containing the user login/signup forms, task creation wizard, and volunteer/donation workspaces.
  * [script.js](file:///home/frank/projects/repos/PamojaBuild/frontend/script.js) (0 bytes) — the client-side JavaScript engine required to interact with the backend APIs.

### **CSS Utility Library vs Vanilla CSS**
* The existing landing page is written using clean, structured Vanilla CSS utilizing CSS variables (custom properties) defined in `:root` inside [css/styles.css](file:///home/frank/projects/repos/PamojaBuild/frontend/css/styles.css).
* No Tailwind CSS or other heavy frameworks are configured. The layout uses standard Flexbox and Grid.

---

## **4. Integration & Development Roadmap**

To successfully integrate the frontend and backend and get the entire project working, execute the following steps in order:

### **Phase 1: Backend Fixes & Enhancements**

1. **Add CORS Middleware**:
   Add a simple middleware function in the backend to enable CORS headers.
   ```go
   func CORSMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           w.Header().Set("Access-Control-Allow-Origin", "*")
           w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
           w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
           if r.Method == "OPTIONS" {
               w.WriteHeader(http.StatusOK)
               return
           }
           next.ServeHTTP(w, r)
       })
   }
   ```
   Wrap the main router handler with this middleware in `cmd/server/main.go`.

2. **Fix JSON Success/Error Wrapping**:
   Refactor the handlers in `auth`, `tasks`, and `volunteers` packages to use standard response wrappers:
   ```go
   func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
   }
   func writeError(w http.ResponseWriter, status int, errMessage string) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(map[string]interface{}{"error": errMessage})
   }
   ```

3. **Implement Task Status Auto-Transitions**:
   Update `volunteers/service.go` so that when a volunteer approval brings the count of approved volunteers to `max_volunteers`, the task status is updated to `'in_progress'`.

4. **Correct Keyholder Passwords in Database Seeding**:
   Update `seedKeyholders()` in `cmd/server/main.go` to compute a valid bcrypt hash:
   ```go
   hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
   ```
   Use this hash in the database seed insertion.

---

### **Phase 2: Frontend Implementation**

1. **Build `tasks.html` (Task Discovery Dashboard)**:
   * Implement a grid containing task cards.
   * Add filters for Kenyan Counties (fetched dynamically from `GET /config/counties` or hardcoded) and Task Status (`open`, `in_progress`, `completed`).
   * Include search capabilities and link cards to details modal/views inside `app.html`.

2. **Build `app.html` (Interactive App Interface)**:
   * **Authentication Views**: Tabs/cards for Registering and Logging in.
   * **Task Poster View**: Form to post tasks (with multipart image upload support), view applications, and approve/reject volunteers.
   * **Volunteer Workspace**: Apply for tasks, submit Lightning payment invoices to receive payouts once approved.
   * **Donations Drawer**: Generate a Lightning invoice for a task, show invoice QR code, and poll for payment confirmation.
   * **Keyholder Signature Console**: Panel for logged-in keyholders to view pending payout requests, inspect transaction details (amounts, volunteers, multisig address), and submit signatures/rejections.

3. **Build `script.js` (Client-Side Logic)**:
   * Write functions to handle User Authentication and save JWT to `localStorage`.
   * Add API fetch requests to backend endpoints with appropriate `Authorization: Bearer <token>` headers.
   * Implement image parsing and submission using `FormData` objects.
   * Wire up donation invoice polling to update payment statuses in real-time.
   * Integrate multisig PSBT signature workflow for keyholders.

---

### **Phase 3: Integration & End-to-End Testing**

1. **Polar Regtest Setup**:
   * Open Polar and create a network with 1 Bitcoin Core node and 2-3 LND nodes.
   * Connect backend to a local LND node by filling in `.env` variables (`LND_HOST`, `LND_MACAROON_HEX`, `LND_TLS_CERT_PATH`).
2. **Task Funding & Payout Broadcast Check**:
   * Generate tasks, apply volunteers, and send donations to generate Lightning invoices.
   * Pay donation invoices via LND node (Polar console).
   * Once volunteers are approved and work is marked complete:
     * Derivation generates the 3-of-5 P2WSH script.
     * Fund the generated P2WSH multisig address via Bitcoin RPC (`generatetoaddress` on Polar).
     * Collect 3 keyholder signatures to merge the PSBT, finalize it, and broadcast it to the regtest network.
     * Verify that volunteers receive their corresponding payouts over Lightning automatically.
