Here's the full feature spec:

---

# **PamojaBuild — Feature Specification v1.0**

---

## **Auth & Users**

**What we're building**

* Register with phone number \+ password  
* Login returns a JWT  
* JWT validation middleware applied to all protected routes  
* User roles: `user` and `keyholder`  
* Keyholders are a fixed, pre-seeded set of trusted community members — not self-assigned

**Edge cases & resolutions**

* A user can be both a regular user and a keyholder — role is a field on the user, not a separate account  
* Keyholders are seeded at deploy time (hardcoded or via a seed script), not created through the app UI for now

**Left for later**

* Guest donations (donate without an account) — for launch, all users must have an account  
* OAuth / phone OTP verification  
* Profile pages, avatars, reputation scores

---

## **Tasks**

**What we're building**

* Any logged-in user can create a task with:  
  * Title  
  * Description  
  * Region (selected from a hardcoded dropdown of Kenya's 47 counties)  
  * Optional free-text location detail (e.g. "near Kibuye Market")  
  * Optional donation goal (sats) — display only, no gates  
  * Max volunteers (required — task poster sets this at creation)  
  * Volunteer mode: `open` or `approval_required`  
* List tasks with filters: status, region, category  
* Single task view returns: full details, total sats donated, volunteer list, funding progress toward goal if set  
* Task status lifecycle: `open → in_progress → pending_verification → completed`  
* Status is updated by the wallet service when payout is released — tasks service does not manage this itself  
* Task creation accepts `multipart/form-data` (not JSON) so an image can be uploaded alongside the other fields  
* Image is optional — task creation works fine without one  
* On upload: save the file to `./static/uploads/tasks/` on the server filesystem, store the relative path (e.g. `static/uploads/tasks/abc123.jpg`) in the `image_path` column on the task record  
* Serve the uploads directory as a static file route: `GET /static/*` — so the frontend can display the image at `http://localhost:8080/static/uploads/tasks/abc123.jpg`  
* Accepted formats: JPEG and PNG only. Validate the MIME type on the backend, reject anything else with a 400  
* Max file size: 5MB. Enforce this on the Go handler using `r.ParseMultipartForm(5 << 20)`  
* If no image is uploaded, `image_path` is NULL in the DB — the single task response includes `image_url: null` and the frontend handles that gracefully  
* 

**Edge cases & resolutions**

* Donation goal is optional and purely cosmetic — tasks can be funded beyond the goal, donations never close automatically  
* Max volunteers is required. If the cap is reached, volunteering closes automatically. Task poster can raise the cap after the fact but cannot lower it (can't remove volunteers who already claimed a spot)  
* Volunteer mode is set at creation: `open` means first-come-first-served up to cap; `approval_required` means poster manually approves each applicant  
* Region is validated against the hardcoded county list on the backend — free text is not accepted for region field  
* Status transitions are only triggered by internal service calls (wallet service), never directly by user-facing API calls

**Left for later**

* Map pins and coordinate-based search — add `latitude` and `longitude` nullable columns in a future migration, region field stays and can be auto-populated via reverse geocoding  
* GPS-based "tasks near me" using device location  
* Task editing after creation  
* Task cancellation and associated refund flow

---

## **Donations**

**What we're building**

* Logged-in user donates to a specific task — triggers Lightning invoice creation  
* Donation record stored with: amount in sats, task ID, donor user ID, payment hash, status (`pending → confirmed`)  
* Invoice payment confirmed via polling `lightning.CheckPaymentStatus()`  
* Endpoint to get total confirmed sats for a task (used for progress bar)  
* Donors can choose whether their name is visible publicly on the task — stored as a boolean `is_anonymous` on the donation record. Always recorded internally, optionally shown publicly

**Edge cases & resolutions**

* Unpaid invoices (user generates invoice but never pays) — store as `pending`, ignore in total calculation, clean up via a background job later  
* Overfunding is allowed — no cap on donations regardless of whether a goal is set  
* Refunds on task failure are a manual process for now (reach out to donor via their account) — automated refund flow is not built at launch

**Left for later**

* Guest donations (no account required)  
* Automated refunds on task failure or rejection  
* Donation expiry — auto-cancel pending invoices after X hours

---

## **Volunteers**

**What we're building**

* Logged-in user can apply to volunteer for an open task  
* If task is `open` mode: volunteer is auto-approved up to the cap, status set to `approved` immediately  
* If task is `approval_required` mode: volunteer status set to `pending`, task poster manually approves or rejects via an endpoint  
* Volunteer statuses: `pending → approved → paid`  
* When payout is released, all `approved` volunteers on that task get an equal share of total confirmed sats, status updated to `paid`  
* Task poster cannot volunteer for their own task

**Edge cases & resolutions**

* Cap reached: volunteering endpoint returns an error, no new applications accepted until poster raises the cap  
* Poster raises cap: reopens volunteering automatically  
* Volunteer drops out: not built at launch — once approved, a volunteer is committed (simplifies payout math)  
* Payout split is equal among all approved volunteers — no weighted splits at launch

**Left for later**

* Volunteer drop-out / withdrawal flow  
* Weighted payouts (e.g. based on hours contributed)  
* Volunteer ratings or reputation after task completion

---

## **Wallet & Multisig**

**What we're building**

* Each task has an associated payout request created when volunteers are ready and the task poster marks work as done  
* Keyholders (3-of-5) each sign or reject a payout request via dedicated endpoints  
* When 3 signatures are collected, payout auto-triggers: `lightning.PayInvoice()` is called for each approved volunteer, sats split equally  
* Payout request statuses: `pending → approved → released` or `rejected`  
* Keyholder records and signatures stored in DB: `keyholders`, `payout_requests`, `payout_signatures` tables  
* Donated sats are held in a real Bitcoin multisig address (P2WSH, M-of-N) on regtest — not just application-layer approval tracking  
* Each keyholder holds an actual private key. When a payout is approved, keyholders sign a PSBT (Partially Signed Bitcoin Transaction). Once the threshold of signatures is collected, the PSBT is finalized and broadcast to the regtest network  
* The app coordinates PSBT construction and collection of partial signatures — enforcement is at the protocol layer, not just the database

**Edge cases & resolutions**

* **Keyholder ghosts / never responds:** if a payout request stays open past 72 hours with at least 2 signatures, a background job auto-releases the payout. Build the job, wire the check — even if the timing is hardcoded for now  
* **Keyholder actively rejects:** a single rejection doesn't block the payout (3 approvals still wins). If 3 keyholders reject, payout is blocked and the task is reopened for new volunteers to attempt — sats stay in the task's allocation, not refunded  
* **What happens to money on failure:** sats stay allocated to the task. Task is reopened. Donors who want a refund must request one manually  
* A keyholder cannot sign the same payout request twice — enforced at the DB level with a unique constraint on `(payout_request_id, keyholder_id)`

**Left for later**

* Automated refunds to donors on permanent task failure  
* Adjustable thresholds per task (right now 3-of-5 is hardcoded)  
* Keyholder rotation / management UI

---

## **Lightning Integration**

**What we're building**

* Thin client in `internal/lightning/client.go` — no business logic  
* Three functions only:  
  * `CreateInvoice(amountSats int64, memo string) (paymentRequest string, paymentHash string, error)`  
  * `PayInvoice(paymentRequest string) error`  
  * `CheckPaymentStatus(paymentHash string) (settled bool, error)`  
* Client talks to a local LND node managed by Polar (regtest environment)  
* LND exposed via its REST API — configure host and macaroon in `config/config.go`  
* Everything else (routing, logic, status updates) lives in the service layers that call these

**Edge cases & resolutions**

* If `PayInvoice` fails for one volunteer, log the error and continue paying the rest — flag that volunteer's status as `pay_failed` for manual resolution, don't block the whole payout  
* Polar manages the regtest Bitcoin node and LND nodes — dev setup is: install Polar, create a network, wire the LND endpoint into your `.env`  
* Your `.env.example` should include:  
* LND\_HOST=localhost:8080  
* LND\_MACAROON\_HEX=\<hex encoded admin macaroon from Polar\>  
* LND\_TLS\_CERT\_PATH=\<path to tls.cert from Polar\>

**Left for later**

* Real-time payment notifications via webhooks instead of polling  
* Keysend payments (pay without a prior invoice — volunteers would need to provide a node pubkey)  
* Mainnet deployment

---

## **Location & Region**

**What we're building**

* Hardcoded list of Kenya's 47 counties lives in `config/counties.go` as a Go slice  
* Task creation validates region against this list  
* `GET /tasks` accepts a `?region=` query param — simple WHERE clause filter  
* Frontend gets the county list from a `GET /config/counties` endpoint (or just hardcodes the same list)

**Left for later**

* Map pins (`latitude`, `longitude` columns) — schema is designed to accept them as nullable columns whenever ready  
* GPS-based filtering ("near me")  
* Sub-county or ward level filtering

---

## **API Summary (for agent prompts)**

| Method | Route | Auth | Description |
| ----- | ----- | ----- | ----- |
| POST | /auth/register | Public | Register with phone \+ password |
| POST | /auth/login | Public | Login, returns JWT |
| GET | /tasks | Public | List tasks, filter by ?region= ?status= ?category= |
| POST | /tasks | Required | Create task |
| GET | /tasks/:id | Public | Get single task with donations total \+ volunteers |
| POST | /tasks/:id/volunteer | Required | Apply to volunteer |
| POST | /tasks/:id/volunteers/:vid/approve | Required (poster) | Approve a volunteer application |
| POST | /tasks/:id/complete | Required (poster) | Mark work done, triggers payout request |
| POST | /donations/:task\_id | Required | Create Lightning invoice for donation |
| GET | /donations/:task\_id/total | Public | Total confirmed sats for a task |
| POST | /wallet/payout/:id/sign | Required (keyholder) | Sign a payout request |
| POST | /wallet/payout/:id/reject | Required (keyholder) | Reject a payout request |
| GET | /config/counties | Public | Get list of valid Kenyan counties |

---

## **DB Tables (quick reference)**

| Table | Key fields |
| ----- | ----- |
| users | id, phone, password\_hash, role, created\_at |
| tasks | id, creator\_id, title, description, region, location\_detail, status, goal\_sats, max\_volunteers, volunteer\_mode, created\_at, image\_path(nullable)  |
| volunteers | id, task\_id, user\_id, status, created\_at |
| donations | id, task\_id, donor\_id, amount\_sats, payment\_hash, payment\_request, is\_anonymous, status, created\_at |
| keyholders | id, user\_id |
| payout\_requests | id, task\_id, total\_sats, status, created\_at |
| payout\_signatures | id, payout\_request\_id, keyholder\_id, action (sign/reject), created\_at |

---

Note that chosen DB is sqlite

---

Implementation order.

---

1. **DB setup** — `db/db.go` connection \+ all migrations. Every table, all columns as specced. This is the foundation; nothing else starts without it.

2. **Config** — `config/config.go` and `.env.example`. JWT secret, LND connection details, server port. The agent will need this wired before anything that touches auth or Lightning.

3. **Auth** — register, login, JWT generation. Then the auth middleware. Verify a protected route rejects a request without a token before moving on.

4. **Counties endpoint** — `GET /config/counties`. Tiny, but tasks validation depends on it and it's a good warm-up before the heavier domains.

5. **Tasks** — full CRUD. Create (multipart/form-data, image handling included), list with filters, single task view. Don't wire up the status transitions yet — those come from wallet later.

6. **Volunteers** — apply, approve/reject, cap enforcement. Depends on tasks existing.

7. **Lightning client** — the three functions in `internal/lightning/client.go` pointed at your Polar LND node. Test each function in isolation before moving on.

8. **Donations** — create invoice (calls lightning), poll for confirmation, total sats endpoint.

9. **Wallet & Multisig** — payout request creation, PSBT construction, keyholder signing, threshold check, payout broadcast. This is the hardest piece — give the agent the most context here and don't rush it.

10. **Status wiring** — wallet service updates task status to `completed` and volunteer statuses to `paid` after successful payout. This is the thread that connects everything together.

