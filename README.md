# PamojaBuild

> Communities funding solutions together.

PamojaBuild is a community-powered action platform built on Bitcoin and the Lightning Network that enables residents to identify local problems, pool donations in sats, coordinate volunteers, and reward people who complete community improvement projects.

Instead of waiting for governments, NGOs, or external organizations to solve everyday issues, communities can organize themselves, fund solutions collectively, and transparently compensate volunteers who take action.

---

## The Problem

Across many communities, there are small but impactful problems that remain unsolved because nobody clearly owns them.

Examples include:

* Overgrown bushes becoming criminal hideouts
* Uncollected garbage in markets
* Unsafe public spaces
* Damaged community infrastructure
* Flood cleanup efforts
* Community safety initiatives

People notice these problems every day, but there is often no simple way to:

* Organize volunteers
* Pool resources
* Coordinate action
* Transparently manage funds
* Reward contributors

---

## The Solution

PamojaBuild creates a community action layer where local problems can be turned into community-funded projects.

### How It Works

1. A resident creates a project and uploads evidence photos.
2. Community members donate sats toward the project.
3. Volunteers join the project.
4. Work is completed.
5. Trusted keyholders approve payout.
6. Volunteers are paid via Lightning Network.
7. Contributors earn recognition through impact rankings.

---

## Why Bitcoin & Lightning?

Traditional payment systems introduce delays, fees, and banking requirements.

Lightning provides:

* Instant payments
* Near-zero transaction fees
* Borderless transactions
* No bank account requirements
* Transparent payment tracking

Volunteers can receive rewards immediately after project completion.

---

# Features

## Community Project Feed

The homepage acts as a live feed of community projects.

Each project contains:

* Project image (required)
* Title
* Description
* County
* Specific location
* Funding progress
* Volunteer information
* Status updates

---

## Donations

Users can support projects by donating Bitcoin through the Lightning Network.

Features:

* Lightning invoice generation
* QR code payments
* Anonymous donations
* Real-time funding progress
* Donation history

---

## Volunteer Coordination

Users can volunteer to help solve community problems.

### Open Mode

Volunteers are automatically accepted.

### Approval Mode

Project creators manually approve volunteers.

---

## Multisig Payout Protection

To protect donated funds, payouts require approval from trusted community keyholders.

### Example

* 5 keyholders exist
* Any 3 approvals are required
* Funds cannot be released by a single individual

This creates transparency and community accountability.

---

## Impact Rankings

Community contributions are publicly recognized through leaderboards.

### Top Donors

Users who have donated the most sats.

### Top Volunteers

Users who have completed the most volunteer work.

### Top Project Leaders

Users who have successfully created and completed projects.

### Overall Impact

A combined ranking based on:

* Donations
* Volunteer work
* Successful projects

---

## User Profiles

Each user profile includes:

* Projects created
* Volunteer history
* Donation history
* Impact score
* Community badges
* Keyholder approvals (if applicable)

---

## Community Badges

### Donation Badges

* 🌱 First Donation
* ⚡ Community Supporter
* 💰 Major Contributor

### Volunteer Badges

* 🤝 First Volunteer
* 🛠 Community Builder
* 🏅 Local Hero

### Project Badges

* 📍 First Project
* 🚀 Community Organizer
* 🏗 Change Maker

---

# Architecture

## Backend

Built using Go.

```
backend/
├── cmd/
├── internal/
│   ├── auth/
│   ├── config/
│   ├── db/
│   ├── models/
│   ├── tasks/
│   ├── volunteers/
│   ├── donations/
│   ├── wallet/
│   ├── lightning/
│   ├── profile/
│   ├── leaderboard/
│   ├── middleware/
│   └── router/
└── scripts/
```

### Backend Architecture

```
HTTP Request
↓
Handler
↓
Service
↓
Repository
↓
Database
```

---

## Frontend

Built using:

* Next.js
* TypeScript
* TailwindCSS
* shadcn/ui
* TanStack Query
* Axios
* React Hook Form
* Zod

### Main Pages

* Home Feed
* Project Details
* Profile
* Impact Rankings
* Login
* Register

---

# Database

Core entities:

```
users
tasks
volunteers
donations
keyholders
payout_requests
payout_signatures
```

Relationships:

```
User
├── Tasks
├── Donations
└── Volunteer Applications

Task
├── Donations
├── Volunteers
└── Payout Request
```

---

# Lightning Integration

The platform connects to an LND node.

Supported operations:

### Create Invoice

Generates Lightning invoices for donations.

### Check Payment Status

Verifies whether a donation has been settled.

### Pay Invoice

Pays volunteers after project completion.

---

# Project Lifecycle

```
Create Project
↓
Receive Donations
↓
Recruit Volunteers
↓
Work Completed
↓
Payout Request Created
↓
Keyholder Approvals
↓
Lightning Payout
↓
Project Completed
```

---

# Security

## Authentication

JWT-based authentication.

## Authorization

Role-based access control.

Roles:

```
user
keyholder
admin (future)
```

## Multisig Protection

Funds cannot be released by a single individual.

A minimum approval threshold is required before payouts can be executed.

---

# Future Roadmap

## Phase 1

* Community project feed
* Donations
* Volunteer management
* Lightning payouts
* Impact rankings

## Phase 2

* Real-time notifications
* Project comments
* Progress updates
* Enhanced reporting

## Phase 3

* Disaster response coordination
* County-level dashboards
* NGO integration
* Reputation system
* Community governance

---

# Local Development

## Backend

```bash
cd backend

cp .env.example .env

go mod tidy

go run cmd/server/main.go
```

## Frontend

```bash
cd frontend

npm install

npm run dev
```

---

# Environment Variables

```env
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=pamojabuild

JWT_SECRET=supersecret

LND_HOST=localhost:10009
LND_MACAROON_HEX=
LND_TLS_CERT_PATH=
```

---

# Vision

PamojaBuild is not a crowdfunding platform.

It is not a task marketplace.

It is not a social network.

PamojaBuild is a community-owned action layer where local problems, local funding, and local people come together to create measurable change.

By combining community coordination with Bitcoin-powered incentives, PamojaBuild enables neighborhoods to solve problems faster, more transparently, and with greater accountability than traditional approaches.

---

## License

MIT License

Copyright (c) 2026 PamojaBuild
