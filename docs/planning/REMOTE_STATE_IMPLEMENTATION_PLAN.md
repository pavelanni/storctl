# storctl Remote State & Cost Control Implementation Plan

**Date:** 2025-11-04
**Status:** Planning/Design Phase
**Timeline:** 8-10 weeks implementation
**Branch:** `feature/remote-state-cost-tracking`

## Executive summary

This document outlines the plan to enhance storctl with remote state storage, cost tracking, automated TTL enforcement, and a web dashboard. The goal is to solve the "runaway VM" problem where cloud instances are forgotten and accumulate costs.

### Problem statement

Current issues with Terraform-based approach:
- State files stored locally on team members' laptops
- No coordination between team members
- Manual deletion of cloud instances leaves DNS records orphaned
- No visibility into current spend by owner/project
- No automated cleanup of forgotten resources

### Solution approach

Build a lightweight web service with PostgreSQL backend that provides:
- Centralized state storage (all team members see same state)
- Web dashboard for cost visibility and management
- Automated TTL enforcement (daemon deletes expired resources)
- Budget tracking and alerts
- Audit logging for compliance
- Enhanced CLI that can work in local or remote mode

## Requirements analysis

Based on team discussion:

- **Timeline:** 8-12 weeks available
- **Dashboard:** Nice to have (not blocking)
- **Approvals:** Not needed initially (can add later)
- **Operations:** Team comfortable running web services
- **Constraints:**
  - Must use Hetzner Cloud for VMs
  - Must use Route53 for DNS (currently Cloudflare in code, needs update)
  - Must use Ansible for VM setup
  - Cannot use AWS S3 (company is MinIO)
  - Can use Hetzner Object Storage (S3-compatible)
  - Small team (handful of people)

## Options considered

### Option A: Enhanced storctl with Hetzner Object Storage

**Pros:** Minimal changes, low risk, fast implementation (4-6 weeks)
**Cons:** S3 lacks native locking, no web dashboard, CLI-only

**Cost:** €4-10/month

### Option B: Switch to Pulumi

**Pros:** Industry standard, strong ecosystem, multi-cloud ready
**Cons:** Complete rewrite, 8-12 weeks, doesn't solve cost control out-of-box, high learning curve

**Cost:** €1-10/month + high development cost

### Option C: Web Service + Dashboard (RECOMMENDED)

**Pros:** Centralized control, web UI, automated enforcement, best management visibility
**Cons:** Operational overhead, more code to write

**Cost:** €7-8/month + 8-10 weeks development

### Option D1: Enhanced storctl + PostgreSQL (no web service)

**Pros:** Better than Option A, good middle ground, can evolve to Option C
**Cons:** No web dashboard initially, still CLI-focused

**Cost:** €3.79-10/month

## Recommended solution: Option C

Based on your requirements (8-12 week timeline, comfortable with operations, dashboard nice-to-have), **Option C: Web Service + Dashboard** is the best fit.

### Why Option C wins

1. **Solves all stated problems** - runaway VMs, cost visibility, state coordination
2. **Timeline fits** - 8-10 weeks matches your availability
3. **Management appeal** - Web dashboard provides visual impact for stakeholders
4. **Team capacity** - You're comfortable operating services
5. **Future-proof** - Can add approval workflows, multi-cloud, etc. later
6. **Cost effective** - €7-8/month infrastructure cost

## Architecture design

### High-level architecture

```
┌──────────────────────────────────────────┐
│  Hetzner CX21 Server (€5.90/month)       │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  storctl-server (Go web service)    │ │
│  │  - REST API                         │ │
│  │  - Authentication (API tokens)      │ │
│  │  - Business logic                   │ │
│  │  Port: 8080 (internal)              │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  PostgreSQL 15+                     │ │
│  │  - Labs state                       │ │
│  │  - Audit logs                       │ │
│  │  - Cost history                     │ │
│  │  - Users/API tokens                 │ │
│  │  Port: 5432 (localhost only)        │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Web UI (served by storctl-server)  │ │
│  │  - Dashboard (React/Vue/Svelte)     │ │
│  │  - Lab management                   │ │
│  │  - Cost reports                     │ │
│  │  Port: 443 (HTTPS via Caddy)        │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  TTL Enforcement Daemon             │ │
│  │  (goroutine in storctl-server)      │ │
│  │  - Checks every 15 minutes          │ │
│  │  - Sends warnings                   │ │
│  │  - Executes cleanup                 │ │
│  └────────────────────────────────────┘ │
└──────────────────────────────────────────┘
            ↑
            │ HTTPS API calls
            │
┌───────────┴──────────────────────────────┐
│  storctl CLI (on dev machines)           │
│  - Can run in local mode (bbolt)         │
│  - Or remote mode (API calls)            │
│  - Config: ~/.storctl/config.yaml        │
└──────────────────────────────────────────┘
            ↓
┌──────────────────────────────────────────┐
│  Hetzner Cloud API                        │
│  - VMs, Volumes, SSH Keys                │
└──────────────────────────────────────────┘
            ↓
┌──────────────────────────────────────────┐
│  Route53 DNS API (or Cloudflare)         │
│  - A records for servers                 │
└──────────────────────────────────────────┘
```

### Database schema

```sql
-- Labs table (main state)
CREATE TABLE labs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    api_version VARCHAR(50) NOT NULL,
    kind VARCHAR(50) NOT NULL,
    metadata JSONB NOT NULL,
    spec JSONB NOT NULL,
    status JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Audit logs
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- create, delete, update, extend
    resource_type VARCHAR(50) NOT NULL, -- lab, server, volume
    resource_name VARCHAR(255) NOT NULL,
    details JSONB,
    ip_address INET
);

-- Cost history
CREATE TABLE cost_history (
    id SERIAL PRIMARY KEY,
    lab_name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    project VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    hourly_cost DECIMAL(10,4) NOT NULL,
    accumulated_cost DECIMAL(10,4) NOT NULL,
    resource_details JSONB
);

-- Users and API tokens
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE api_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Budget configuration
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    owner VARCHAR(255),
    project VARCHAR(255),
    monthly_limit DECIMAL(10,2) NOT NULL,
    alert_threshold INTEGER DEFAULT 80, -- percentage
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT budget_target_check CHECK (
        (owner IS NOT NULL AND project IS NULL) OR
        (owner IS NULL AND project IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_labs_name ON labs(name);
CREATE INDEX idx_labs_deleted_at ON labs(deleted_at);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_name);
CREATE INDEX idx_cost_history_lab ON cost_history(lab_name);
CREATE INDEX idx_cost_history_owner ON cost_history(owner);
CREATE INDEX idx_cost_history_timestamp ON cost_history(timestamp);
```

### REST API specification

```
Base URL: https://storctl.yourdomain.com/api/v1
Authentication: Bearer token in Authorization header

Endpoints:

POST   /auth/token              Generate API token (admin only initially)
GET    /auth/token              List user's tokens
DELETE /auth/token/:id          Revoke token

GET    /labs                    List all labs (supports ?owner=X&project=Y)
POST   /labs                    Create lab (YAML in body)
GET    /labs/:name              Get lab details
DELETE /labs/:name              Delete lab (?force=true)
POST   /labs/:name/extend       Extend TTL (body: {ttl: "4h"})

GET    /servers                 List servers (supports ?lab=X)
GET    /servers/:name           Get server details
DELETE /servers/:name           Delete server

GET    /volumes                 List volumes
GET    /volumes/:name           Get volume details
DELETE /volumes/:name           Delete volume

GET    /costs                   Get cost summary (?owner=X&project=Y&since=DATE)
GET    /costs/labs/:name        Get specific lab costs
GET    /costs/projection        Get monthly projection

GET    /audit                   Get audit logs (?since=DATE&user=X&action=Y)

GET    /budgets                 List budgets
POST   /budgets                 Create budget
PUT    /budgets/:id             Update budget
DELETE /budgets/:id             Delete budget

GET    /health                  Health check
GET    /metrics                 Prometheus metrics (optional)
```

### Configuration updates

Update `~/.storctl/config.yaml`:

```yaml
# Storage backend configuration
storage:
  type: remote  # or "local" for backward compatibility
  url: https://storctl.yourdomain.com
  api_token: your-api-token-here

# Legacy local storage (kept for backward compatibility)
# storage:
#   type: local
#   path: ~/.storctl/labs.db
#   bucket: labs

providers:
  - name: hetzner
    token: your-hetzner-token
    location: nbg1
  - name: lima

dns:
  provider: route53  # CHANGED from cloudflare
  access_key_id: your-aws-access-key
  secret_access_key: your-aws-secret-key
  hosted_zone_id: your-zone-id
  domain: yourdomain.com

# Budget alerts (validated on server, not client)
budgets:
  - owner: pavel
    monthly_limit: 100
    alert_at: 80
  - project: aistor
    monthly_limit: 500
    alert_at: 90

# Notification settings
notifications:
  email:
    enabled: true
    smtp_server: smtp.gmail.com:587
    from: storctl@yourdomain.com
    recipients:
      - pavel@yourdomain.com
  slack:
    enabled: false
    webhook_url: https://hooks.slack.com/...

email: your-email@example.com
organization: your-organization
owner: your-name
```

Server configuration (`/etc/storctl-server/config.yaml`):

```yaml
server:
  host: 0.0.0.0
  port: 8080
  tls: false  # Caddy handles TLS

database:
  host: localhost
  port: 5432
  user: storctl
  password: ${POSTGRES_PASSWORD}  # from environment
  database: storctl
  ssl_mode: disable

auth:
  token_expiry: 8760h  # 1 year

daemon:
  ttl_check_interval: 15m
  warning_intervals:
    - 24h
    - 4h
    - 1h

cost_tracking:
  update_interval: 1h
  pricing:
    hetzner:
      servers:
        cx11: 3.79
        cx21: 5.90
        cx22: 6.90
        # ... add all types
      volumes_per_gb: 0.044

notifications:
  email:
    enabled: true
    smtp_server: smtp.gmail.com:587
    from: storctl@yourdomain.com
```

## Implementation plan

### Phase 1: Backend infrastructure (Weeks 1-3)

#### Week 1: Database & storage abstraction

**Tasks:**

1. Create database schema
   - Write migration SQL files
   - Add migration tooling (golang-migrate or similar)
   - Create database setup scripts

1. Add storage interface abstraction
   - Create `internal/storage/interface.go` with Storage interface
   - Move existing bbolt code to `internal/storage/local/`
   - Implement PostgreSQL storage in `internal/storage/postgres/`
   - Add factory function to select backend based on config

1. Migration tool
   - Command: `storctl migrate --from local --to remote`
   - Read labs from bbolt
   - Upload to PostgreSQL via API
   - Verify migration success

1. Configuration updates
   - Add storage backend selection to config.yaml
   - Add remote URL and token fields
   - Update config parsing in `internal/config/`

**Files to create/modify:**
- `internal/storage/interface.go` (new)
- `internal/storage/local/bbolt.go` (move from internal/lab/lab.go)
- `internal/storage/postgres/postgres.go` (new)
- `internal/storage/factory.go` (new)
- `migrations/001_initial_schema.up.sql` (new)
- `migrations/001_initial_schema.down.sql` (new)
- `cmd/migrate.go` (new)
- `internal/config/config.go` (modify)

**Testing:**
- Unit tests for PostgreSQL storage operations
- Integration tests with test database
- Migration tool validation

#### Week 2: Web service foundation

**Tasks:**

1. Create web service binary
   - New main package: `cmd/storctl-server/main.go`
   - HTTP server setup with Echo/Gin framework
   - Graceful shutdown handling
   - Configuration loading

1. Core API endpoints
   - POST /api/v1/labs (create lab)
   - GET /api/v1/labs (list labs)
   - GET /api/v1/labs/:name (get lab)
   - DELETE /api/v1/labs/:name (delete lab)
   - POST /api/v1/labs/:name/extend (extend TTL)

1. Authentication middleware
   - API token validation
   - Token storage in database
   - Token generation endpoint (admin only)

1. Database connection management
   - Connection pooling
   - Health checks
   - Retry logic

1. Error handling and logging
   - Structured logging (using existing logger)
   - HTTP error responses (consistent format)
   - Request ID tracking

**Files to create:**
- `cmd/storctl-server/main.go` (new)
- `internal/server/server.go` (new)
- `internal/server/handlers/labs.go` (new)
- `internal/server/handlers/auth.go` (new)
- `internal/server/middleware/auth.go` (new)
- `internal/server/middleware/logging.go` (new)
- `internal/server/routes.go` (new)

**Testing:**
- API endpoint tests with httptest
- Authentication tests
- Error handling tests

#### Week 3: CLI client updates

**Tasks:**

1. Create API client package
   - HTTP client with retry logic
   - Token management
   - API method implementations matching REST endpoints

1. Update CLI commands to support remote mode
   - Check config for storage type (local vs remote)
   - Route to appropriate backend
   - Maintain backward compatibility with local mode

1. New CLI commands
   - `storctl config set-remote --url URL --token TOKEN`
   - `storctl config set-local`
   - `storctl auth token create --name NAME`
   - `storctl auth token list`
   - `storctl auth token revoke --id ID`

1. Error handling improvements
   - Better error messages for API failures
   - Network timeout handling
   - Token expiry detection

**Files to create/modify:**
- `internal/client/client.go` (new)
- `internal/client/labs.go` (new)
- `internal/client/auth.go` (new)
- `cmd/config.go` (modify)
- `cmd/create_lab.go` (modify)
- `cmd/get_lab.go` (modify)
- `cmd/delete_lab.go` (modify)
- All other resource commands (modify)

**Testing:**
- Mock server tests for client
- Integration tests with local server
- Backward compatibility tests with local mode

### Phase 2: Cost tracking & TTL enforcement (Weeks 4-6)

#### Week 4: Cost calculation engine

**Tasks:**

1. Create cost calculation package
   - Hetzner pricing data (servers, volumes, IPs)
   - Cost calculation functions
   - Hourly and monthly projections

1. Track costs during lab lifecycle
   - Record creation timestamp
   - Calculate costs on query
   - Store cost snapshots in cost_history table

1. API endpoints for cost data
   - GET /api/v1/costs (aggregate costs)
   - GET /api/v1/costs/labs/:name (lab-specific)
   - GET /api/v1/costs/projection (monthly projection)

1. CLI commands
   - `storctl costs show [--owner X] [--project Y] [--since DATE]`
   - `storctl costs projection`
   - Enhanced `storctl get lab` to show current cost

**Files to create/modify:**
- `internal/cost/pricing.go` (new)
- `internal/cost/calculator.go` (new)
- `internal/server/handlers/costs.go` (new)
- `cmd/costs.go` (new)
- `cmd/get_lab.go` (modify - add cost display)

**Testing:**
- Cost calculation accuracy tests
- Projection tests
- API endpoint tests

#### Week 5: Background daemon

**Tasks:**

1. TTL enforcement daemon
   - Goroutine runs in storctl-server
   - Ticker checks every 15 minutes
   - Query labs with expired TTLs
   - Calculate warning times (24h, 4h, 1h before expiry)

1. Notification system
   - Email sender (SMTP)
   - Slack webhook sender
   - Notification templates
   - Track sent notifications (don't spam)

1. Safe cleanup implementation
   - Pre-deletion validation
   - Delete resources via provider interface
   - Audit logging
   - Rollback on errors

1. Daemon management
   - Start/stop with server
   - Status endpoint: GET /api/v1/daemon/status
   - Manual trigger: POST /api/v1/daemon/cleanup (admin only)

**Files to create:**
- `internal/daemon/daemon.go` (new)
- `internal/daemon/ttl.go` (new)
- `internal/daemon/notifications.go` (new)
- `internal/notifications/email.go` (new)
- `internal/notifications/slack.go` (new)
- `internal/server/handlers/daemon.go` (new)

**Testing:**
- Mock provider for safe testing
- Notification delivery tests
- TTL calculation tests
- Cleanup logic tests

#### Week 6: Budget & alerting

**Tasks:**

1. Budget management
   - Database schema for budgets (already in Phase 1)
   - API endpoints: GET/POST/PUT/DELETE /api/v1/budgets
   - Budget validation during lab creation

1. Alert generation
   - Check budgets during cost updates
   - Send alerts when threshold exceeded
   - Track alert history (don't spam)
   - Alert types: warning (80%), critical (95%), exceeded (100%)

1. CLI commands
   - `storctl budget list`
   - `storctl budget create --owner X --limit Y`
   - `storctl budget delete --id Z`

1. Budget enforcement (optional)
   - Block lab creation if budget exceeded
   - Override flag: `--force-over-budget` (admin only)

**Files to create/modify:**
- `internal/server/handlers/budgets.go` (new)
- `internal/budget/budget.go` (new)
- `internal/budget/alerts.go` (new)
- `cmd/budget.go` (new)
- `internal/server/handlers/labs.go` (modify - add budget check)

**Testing:**
- Budget validation tests
- Alert triggering tests
- Override logic tests

### Phase 3: Web dashboard (Weeks 7-9)

#### Week 7: Dashboard foundation

**Tasks:**

1. Frontend setup
   - Choose framework (React + Vite recommended for MinIO familiarity)
   - Project structure: `web/`
   - Build system integration
   - Embed built assets in Go binary (embed.FS)

1. Authentication UI
   - Login page
   - Token storage (localStorage)
   - Logout functionality
   - Protected routes

1. Layout and navigation
   - Top navigation bar
   - Sidebar menu
   - Dashboard cards/widgets
   - Responsive design

1. API integration
   - Axios/fetch client setup
   - Authentication interceptor
   - Error handling
   - Loading states

**Files to create:**
- `web/package.json` (new)
- `web/src/main.jsx` (new)
- `web/src/App.jsx` (new)
- `web/src/components/Login.jsx` (new)
- `web/src/components/Layout.jsx` (new)
- `web/src/api/client.js` (new)
- `cmd/storctl-server/main.go` (modify - serve frontend)

#### Week 8: Core dashboard views

**Tasks:**

1. Labs list view
   - Table with columns: Name, Owner, Status, TTL, Cost, Actions
   - Filters: owner, project, status
   - Sort by: name, created, cost, expiry
   - Actions: View, Delete, Extend TTL

1. Lab details page
   - Lab metadata and spec
   - Server list with details
   - Volume list with details
   - Cost breakdown
   - Timeline (created, expires)
   - Extend TTL form

1. Cost dashboard
   - Total current spend (gauge)
   - Monthly projection (number)
   - Cost by owner (pie chart)
   - Cost by project (bar chart)
   - Cost trend over time (line chart)
   - Top 5 expensive labs (table)

1. Charts and visualizations
   - Use Chart.js or Recharts
   - Real-time updates (polling or WebSocket)

**Files to create:**
- `web/src/pages/LabsList.jsx` (new)
- `web/src/pages/LabDetails.jsx` (new)
- `web/src/pages/CostDashboard.jsx` (new)
- `web/src/components/LabsTable.jsx` (new)
- `web/src/components/CostCharts.jsx` (new)

#### Week 9: Management features

**Tasks:**

1. Audit log viewer
   - Table with: timestamp, user, action, resource, details
   - Filters: date range, user, action, resource
   - Export to CSV
   - Search functionality

1. Budget management UI
   - List budgets (table)
   - Create budget (modal/form)
   - Edit budget
   - Delete budget
   - Visual indicators (progress bars)

1. Lab creation form (optional)
   - YAML editor
   - Form-based creation
   - Template selection
   - Validation

1. User management (basic)
   - List API tokens
   - Create new token
   - Revoke token
   - Show last used

**Files to create:**
- `web/src/pages/AuditLogs.jsx` (new)
- `web/src/pages/Budgets.jsx` (new)
- `web/src/pages/CreateLab.jsx` (new, optional)
- `web/src/pages/Settings.jsx` (new)

### Phase 4: Polish & deployment (Week 10)

#### Week 10: Production ready

**Tasks:**

1. Deployment automation
   - Terraform or Ansible for server setup
   - Install script: PostgreSQL, Caddy, storctl-server
   - Systemd service files
   - Caddy configuration for HTTPS
   - Backup scripts

1. Monitoring setup
   - Health check endpoint
   - Prometheus metrics (optional)
   - Log aggregation setup
   - Alerting for service down

1. Backup and recovery
   - PostgreSQL backup script (pg_dump)
   - Cron job for daily backups
   - Backup to Hetzner Object Storage
   - Recovery documentation and testing

1. Documentation
   - Installation guide
   - User guide (CLI and web)
   - API documentation
   - Operations runbook
   - Architecture documentation

1. Testing and validation
   - End-to-end testing
   - Load testing
   - Security review
   - User acceptance testing

**Files to create:**
- `deployment/terraform/` or `deployment/ansible/` (new)
- `deployment/systemd/storctl-server.service` (new)
- `deployment/caddy/Caddyfile` (new)
- `deployment/backup.sh` (new)
- `docs/INSTALLATION.md` (new)
- `docs/USER_GUIDE.md` (new)
- `docs/API.md` (new)
- `docs/OPERATIONS.md` (new)

## Technical specifications

### Technology stack

**Backend:**
- Language: Go 1.23+
- Web framework: Echo or Gin (recommend Echo for built-in middleware)
- Database: PostgreSQL 15+
- Database driver: pgx (best performance) or lib/pq
- Migrations: golang-migrate
- Testing: testify, httptest

**Frontend:**
- Framework: React 18+ with Vite
- UI library: Tailwind CSS + shadcn/ui or Ant Design
- Charts: Recharts or Chart.js
- HTTP client: Axios
- State management: React Context (or Zustand if needed)
- Build: Vite (fast, modern)

**Infrastructure:**
- Server: Hetzner CX21 (4GB RAM, 2 vCPUs)
- Reverse proxy: Caddy (automatic HTTPS)
- Database: PostgreSQL on same server
- Backups: Hetzner Object Storage

### Code organization

```
storctl/
├── cmd/
│   ├── storctl/              # Main CLI (existing)
│   ├── storctl-server/       # Web service (new)
│   │   └── main.go
│   └── migrate.go            # Migration tool (new)
├── internal/
│   ├── storage/              # Storage abstraction (new)
│   │   ├── interface.go
│   │   ├── factory.go
│   │   ├── local/            # bbolt implementation
│   │   │   └── bbolt.go
│   │   └── postgres/         # PostgreSQL implementation
│   │       └── postgres.go
│   ├── client/               # API client (new)
│   │   ├── client.go
│   │   ├── labs.go
│   │   └── auth.go
│   ├── server/               # Web service (new)
│   │   ├── server.go
│   │   ├── routes.go
│   │   ├── handlers/
│   │   │   ├── labs.go
│   │   │   ├── auth.go
│   │   │   ├── costs.go
│   │   │   ├── budgets.go
│   │   │   └── audit.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       └── logging.go
│   ├── daemon/               # Background tasks (new)
│   │   ├── daemon.go
│   │   ├── ttl.go
│   │   └── notifications.go
│   ├── cost/                 # Cost tracking (new)
│   │   ├── pricing.go
│   │   └── calculator.go
│   ├── budget/               # Budget management (new)
│   │   ├── budget.go
│   │   └── alerts.go
│   ├── notifications/        # Notification system (new)
│   │   ├── email.go
│   │   └── slack.go
│   ├── lab/                  # Lab manager (existing)
│   ├── provider/             # Providers (existing)
│   ├── config/               # Config (existing, modified)
│   └── ...                   # Other existing packages
├── web/                      # Frontend (new)
│   ├── package.json
│   ├── vite.config.js
│   ├── src/
│   │   ├── main.jsx
│   │   ├── App.jsx
│   │   ├── api/
│   │   │   └── client.js
│   │   ├── pages/
│   │   │   ├── LabsList.jsx
│   │   │   ├── LabDetails.jsx
│   │   │   ├── CostDashboard.jsx
│   │   │   ├── AuditLogs.jsx
│   │   │   └── Budgets.jsx
│   │   └── components/
│   │       ├── Layout.jsx
│   │       ├── Login.jsx
│   │       └── ...
│   └── dist/                 # Built assets (embedded in Go binary)
├── migrations/               # Database migrations (new)
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   └── ...
├── deployment/               # Deployment configs (new)
│   ├── terraform/            # or ansible/
│   ├── systemd/
│   │   └── storctl-server.service
│   ├── caddy/
│   │   └── Caddyfile
│   └── backup.sh
└── docs/                     # Documentation
    ├── INSTALLATION.md       # New
    ├── USER_GUIDE.md         # New
    ├── API.md                # New
    └── OPERATIONS.md         # New
```

### Security considerations

1. **Authentication:**
   - API tokens stored as hashed values (bcrypt)
   - Token expiry (default 1 year, configurable)
   - Rate limiting on auth endpoints

1. **Authorization:**
   - Initial implementation: all authenticated users have equal access
   - Future: Role-based access control (admin, user, viewer)

1. **HTTPS:**
   - Caddy provides automatic HTTPS with Let's Encrypt
   - Force HTTPS redirect
   - HSTS headers

1. **Input validation:**
   - Validate all API inputs
   - Sanitize YAML/JSON payloads
   - Prevent SQL injection (use parameterized queries)

1. **Secrets management:**
   - Database password from environment variable
   - API tokens never logged
   - Hetzner/AWS tokens stored encrypted (future enhancement)

1. **Network security:**
   - PostgreSQL listens on localhost only
   - Server firewall (only 80, 443 open)
   - Consider VPN for admin access (future)

### Performance considerations

1. **Database:**
   - Connection pooling (10-20 connections)
   - Indexes on frequently queried columns
   - Periodic VACUUM for PostgreSQL

1. **API:**
   - Pagination for list endpoints (default 50, max 100)
   - Caching for cost calculations (1 hour TTL)
   - Rate limiting (100 requests/minute per token)

1. **Frontend:**
   - Code splitting (React lazy loading)
   - Optimize bundle size (target <500KB)
   - Progressive loading (show data as it arrives)

1. **Cost tracking:**
   - Update costs hourly (not on every request)
   - Store snapshots to avoid recalculation
   - Background job for cost updates

### Monitoring and observability

1. **Logging:**
   - Structured JSON logs
   - Log levels: debug, info, warn, error
   - Request ID tracking through stack
   - Rotate logs daily (logrotate)

1. **Metrics (optional):**
   - Prometheus endpoint: /metrics
   - Metrics: HTTP requests, DB connections, daemon runs, cost totals
   - Grafana dashboard

1. **Health checks:**
   - Endpoint: GET /health
   - Checks: DB connection, provider API reachable
   - Used by monitoring and load balancer

1. **Alerts:**
   - Server down (via external monitoring)
   - Daemon failures
   - Database connection failures
   - Disk space low

## Cost analysis

### Infrastructure costs

**Production server (Hetzner CX21):**
- 2 vCPUs, 4GB RAM, 40GB SSD
- Cost: €5.90/month

**Database:**
- PostgreSQL on same server: €0
- Alternative: Managed database (future): ~€10-20/month

**Domain & SSL:**
- Domain: ~€10-15/year (€1.25/month)
- SSL: Free (Let's Encrypt via Caddy)

**Backups:**
- Hetzner Object Storage: ~€0.50-1/month (50-100GB)

**Total monthly cost: ~€7-8/month**

### Development costs

**Time investment:**
- 8-10 weeks @ 40 hours/week = 320-400 hours
- At internal rate (example €50/hour): €16,000-20,000
- Opportunity cost: other projects delayed

### ROI calculation

**Cost savings:**
- One forgotten lab (3 servers, 8 volumes): €46.57/month
- If prevents 1 forgotten lab per month: Break even in 5 months
- If prevents 2-3 forgotten labs per quarter: ROI in 3 months

**Non-monetary benefits:**
- Better visibility for management
- Compliance (audit trail)
- Team productivity (no manual tracking)
- Professional image (vs. ad-hoc scripts)

## Risks and mitigation

### Technical risks

1. **Risk:** PostgreSQL single point of failure
   **Mitigation:** Daily backups, quick restore procedure, consider replication if critical

1. **Risk:** API breaking changes affect CLI
   **Mitigation:** API versioning (/api/v1), maintain backward compatibility, version negotiation

1. **Risk:** TTL daemon fails and doesn't clean up
   **Mitigation:** Monitoring alerts, manual cleanup command, daemon status checks

1. **Risk:** Cost calculations incorrect
   **Mitigation:** Thorough testing, reconcile with Hetzner bills, manual override

1. **Risk:** Database migration fails
   **Mitigation:** Test thoroughly, backup before migration, rollback plan

### Operational risks

1. **Risk:** Server downtime = no lab management
   **Mitigation:** CLI local mode fallback (read-only), monitoring, quick recovery

1. **Risk:** Data loss
   **Mitigation:** Daily backups to Object Storage, test restore, consider replication

1. **Risk:** Security breach
   **Mitigation:** HTTPS, token expiry, rate limiting, security audits, minimal permissions

### Project risks

1. **Risk:** Timeline slips
   **Mitigation:** Prioritize core features (state, TTL, costs), defer nice-to-haves (approvals, advanced UI)

1. **Risk:** Team capacity changes
   **Mitigation:** Good documentation, modular design, incremental rollout

1. **Risk:** Requirements change
   **Mitigation:** Flexible architecture, API-first design, progressive enhancement

## Success metrics

### Launch criteria (MVP)

Must have:
- ✅ Remote state storage (PostgreSQL)
- ✅ API working (core endpoints)
- ✅ CLI can create/list/delete labs via API
- ✅ TTL daemon runs and cleans up expired labs
- ✅ Basic cost tracking works
- ✅ Web dashboard shows labs and costs
- ✅ Documentation complete

Nice to have (can defer):
- Budget alerts
- Approval workflows
- Advanced charts
- User management

### Success indicators (post-launch)

**Week 1:**
- All team members migrated to remote mode
- No runaway VMs (daemon working)
- Cost visibility dashboard accessed by management

**Month 1:**
- Zero forgotten VMs accumulating costs
- All labs have TTLs set
- Audit logs show all activity

**Month 3:**
- Cost savings measurable (compare to previous quarter)
- Team satisfaction (faster, easier)
- Management confidence (visibility, control)

### Metrics to track

- Number of labs created/deleted per week
- Average lab lifetime
- Cost per owner/project
- Number of TTL extensions
- Daemon cleanup success rate
- API response times
- Dashboard page views

## Migration strategy

### Phase 1: Deploy server (Week 10)

1. Provision Hetzner server
1. Install PostgreSQL, Caddy, storctl-server
1. Configure HTTPS
1. Create initial admin token
1. Test API manually

### Phase 2: Parallel operation (Week 11)

1. Keep current Terraform/manual process running
1. Team members test CLI in remote mode
1. Create new labs via storctl (both local and remote)
1. Verify daemon works
1. Collect feedback

### Phase 3: Migration (Week 12)

1. Export existing cloud resources
1. Import into storctl (as existing labs)
1. Migrate team members to remote mode
1. Decommission Terraform state files

### Phase 4: Full adoption (Week 13+)

1. All new labs via storctl only
1. Monitor cost savings
1. Add remaining features (budgets, etc.)
1. Train new team members

## Next steps

### Immediate actions

1. **Review this plan** with team and stakeholders
1. **Get management approval** for timeline and budget
1. **Set up development environment:**
   - PostgreSQL local instance
   - Test Hetzner account (if not using prod)
1. **Create project board** (GitHub Issues or similar) with tasks
1. **Start Phase 1:** Database schema and storage abstraction

### Decision points

Before starting implementation, decide:

1. **DNS provider:** Stick with Cloudflare or migrate to Route53?
   - Cloudflare: Already implemented, working
   - Route53: Requirement stated in constraints
   - **Recommendation:** Start with Cloudflare, add Route53 support later

1. **Web framework:** Echo or Gin?
   - Echo: Better middleware, built-in validation
   - Gin: Faster performance, larger community
   - **Recommendation:** Echo for middleware richness

1. **Frontend framework:** React, Vue, or Svelte?
   - React: Team familiarity (MinIO uses React)
   - Vue: Simpler, faster learning
   - Svelte: Smallest bundle, best performance
   - **Recommendation:** React (team familiarity)

1. **Deployment:** Terraform or Ansible?
   - Terraform: Declarative, idempotent
   - Ansible: Already used for VM setup
   - **Recommendation:** Ansible (consistency with existing workflow)

### Questions to resolve

1. Domain name for web service? (e.g., storctl.aistorlabs.com)
1. Initial admin user credentials?
1. Email SMTP server details?
1. Slack webhook for notifications?
1. Backup retention policy? (30 days? 90 days?)

## Conclusion

This plan provides a comprehensive roadmap to enhance storctl with remote state storage, cost control, and management visibility. The solution addresses all stated problems (runaway VMs, state coordination, cost tracking) while fitting within your timeline and operational constraints.

**Key benefits:**
- Solves runaway VM problem with automated TTL enforcement
- Provides centralized state storage (team coordination)
- Delivers cost visibility via web dashboard
- Low operational cost (~€7-8/month)
- Reasonable timeline (8-10 weeks)
- Built on solid foundations (PostgreSQL, Go, React)

**Recommended approach:**
1. Start with Phase 1 (backend infrastructure)
1. Get core functionality working (state + API)
1. Add TTL daemon early (solves main problem)
1. Iterate on dashboard (progressive enhancement)
1. Defer nice-to-haves (approvals, advanced features)

This is a **living document** - update it as implementation progresses, requirements change, or new insights emerge.

---

**Document version:** 1.0
**Last updated:** 2025-11-04
**Next review:** After Phase 1 completion
