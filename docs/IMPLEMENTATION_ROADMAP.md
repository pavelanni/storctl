# storctl Implementation Roadmap

**Date:** 2025-11-04
**Timeline:** 10-12 weeks
**Goal:** Replace minio-lab-terraform with enhanced storctl + web service
**Branch:** `feature/remote-state-cost-tracking`

## Quick overview

This roadmap combines two objectives:
1. **Migrate from Terraform** - Replace minio-lab-terraform with storctl
2. **Add web service** - Remote state, cost tracking, TTL automation, dashboard

**Approach:** Incremental delivery - each phase provides usable improvements.

## Week-by-week plan

### Week 1: Foundation + Critical Terraform gaps

**Goal:** Set up database + add Route53 DNS + cloud-init

**Infrastructure tasks:**
- Create PostgreSQL schema and migration scripts
- Add storage abstraction layer (local BoltDB + PostgreSQL)
- Database connection management

**Terraform parity tasks:**
- Implement Route53 DNS provider (`internal/dns/route53.go`)
- Add wildcard DNS record support (both Route53 and Cloudflare)
- Add UserData/cloud-init support to ServerCreateOpts
- Create disk partitioning template (from terraform user_data.tmpl)

**Deliverable:** Database ready + storctl can create VMs with Route53 + disk partitioning

**Testing checklist:**
- [ ] PostgreSQL migrations run successfully
- [ ] Can create lab with Route53 DNS
- [ ] Wildcard records (`*.server.domain`) created
- [ ] VM boots with partitioned disk (`/mnt/data` mounted)

### Week 2: Web service + Ansible enhancements

**Goal:** REST API working + Ansible receives all variables

**Web service tasks:**
- Create `cmd/storctl-server/` web service binary
- REST API endpoints: POST/GET/DELETE /api/v1/labs
- Authentication middleware (API tokens)
- Health checks and logging

**Ansible tasks:**
- Enhanced variable passing (generate group_vars/all.yml)
- Support 50+ configuration variables from lab spec
- Template rendering for Ansible configs
- Training lab mode flag passing

**Deliverable:** API can create labs + Ansible runs with full configuration

**Testing checklist:**
- [ ] API endpoints work (can create/get/delete labs via curl)
- [ ] API token authentication working
- [ ] Ansible receives all variables from lab spec
- [ ] Training lab mode flag affects deployment time

### Week 3: CLI client + deployment orchestration

**Goal:** CLI works with remote API + smart deployment flow

**CLI tasks:**
- API client package (`internal/client/`)
- Update CLI commands to support remote mode
- Config: `storctl config set-remote --url --token`
- Backward compatibility with local mode

**Orchestration tasks:**
- SSH readiness polling (replace fixed 60s wait)
- Sequential server creation for dependencies
- Progress reporting during deployment
- Error handling and partial rollback

**Deliverable:** `storctl create lab mylab` works end-to-end via API

**Testing checklist:**
- [ ] CLI can switch between local and remote modes
- [ ] Remote lab creation works
- [ ] SSH readiness detected before Ansible runs
- [ ] Progress visible during deployment
- [ ] Errors handled gracefully

### Week 4: Cost tracking

**Goal:** Cost calculation and reporting

**Tasks:**
- Cost calculation engine with Hetzner pricing
- Track costs during lab lifecycle
- API endpoints: GET /api/v1/costs
- CLI: `storctl costs show [--owner X]`
- Enhanced `storctl get lab` to show current cost

**Deliverable:** Cost visibility for all labs and owners

**Testing checklist:**
- [ ] Cost calculation accurate for servers/volumes
- [ ] Costs tracked over time
- [ ] API returns cost data
- [ ] CLI shows costs in human-readable format

### Week 5: TTL daemon + access information

**Goal:** Automated cleanup + generated access scripts

**Daemon tasks:**
- Background goroutine for TTL enforcement
- Check every 15 minutes for expired labs
- Notification system (email/Slack)
- Pre-deletion warnings (24h, 4h, 1h)
- Safe cleanup with audit logging

**Access info tasks:**
- Generate SSH commands for all servers
- Service URLs (MinIO, VSCode, HAProxy) with wildcards
- MC client setup scripts
- Site replication scripts
- Store in `~/.storctl/labs/{lab_name}/scripts/`

**Deliverable:** Labs auto-delete + easy access via generated scripts

**Testing checklist:**
- [ ] Daemon runs and checks TTLs
- [ ] Expired labs deleted automatically
- [ ] Warnings sent before deletion
- [ ] Access scripts generated correctly
- [ ] `storctl info lab mylab` shows all access details

### Week 6: Budget management + multi-server coordination

**Goal:** Budget alerts + proper multi-server labs

**Budget tasks:**
- Budget API endpoints (CRUD operations)
- Alert generation when thresholds exceeded
- CLI: `storctl budget list/create/delete`
- Track alert history (avoid spam)

**Multi-server tasks:**
- Server roles (control-plane, worker)
- Creation order enforcement
- Dependency tracking
- Wait for readiness between servers

**Deliverable:** Budget control + complex multi-server labs work

**Testing checklist:**
- [ ] Budget alerts trigger correctly
- [ ] Multi-server labs create in proper order
- [ ] Server roles passed to Ansible correctly
- [ ] Dependencies handled (e.g., control plane before workers)

### Week 7: Dashboard foundation

**Goal:** Web UI basics

**Tasks:**
- Frontend setup (React + Vite)
- Authentication UI (login page)
- Layout and navigation
- Embed built assets in Go binary
- API integration client

**Deliverable:** Can log into web dashboard

**Testing checklist:**
- [ ] Frontend builds successfully
- [ ] Login works with API token
- [ ] Assets embedded and served by Go binary
- [ ] Navigation between pages works

### Week 8: Dashboard core views

**Goal:** Labs and cost views

**Tasks:**
- Labs list view (table with filters/sorting)
- Lab details page (servers, volumes, cost breakdown)
- Cost dashboard (charts and projections)
- Real-time updates (polling)

**Deliverable:** Can view and manage labs via web UI

**Testing checklist:**
- [ ] Labs list shows all active labs
- [ ] Can filter by owner/project
- [ ] Lab details page complete
- [ ] Cost dashboard shows accurate data
- [ ] Charts render correctly

### Week 9: Dashboard management features

**Goal:** Audit logs and budgets in UI

**Tasks:**
- Audit log viewer (filterable table)
- Budget management UI (list/create/edit/delete)
- Lab creation form (optional)
- User/token management
- Export features (CSV downloads)

**Deliverable:** Complete web management interface

**Testing checklist:**
- [ ] Audit logs visible and filterable
- [ ] Budget management works
- [ ] Can create labs via web UI (if implemented)
- [ ] Export to CSV works

### Week 10: Deployment and documentation

**Goal:** Production-ready deployment

**Tasks:**
- Deployment automation (Ansible for server setup)
- Systemd service files
- Caddy configuration (HTTPS)
- Backup scripts (PostgreSQL to Object Storage)
- Installation documentation
- User guide (CLI + web)
- API documentation
- Operations runbook

**Deliverable:** Can deploy to production server + complete docs

**Testing checklist:**
- [ ] Server deployment automated
- [ ] HTTPS works via Caddy
- [ ] Backups run successfully
- [ ] All documentation complete
- [ ] End-to-end testing passed

### Weeks 11-12: Migration and polish

**Goal:** Team migration from Terraform + refinement

**Tasks:**
- Team training on storctl
- Migrate existing knowledge to storctl workflows
- Performance testing and optimization
- Security review
- Bug fixes from real usage
- Collect feedback and iterate

**Deliverable:** Team fully migrated to storctl

**Testing checklist:**
- [ ] All team members using storctl
- [ ] No Terraform deployments for 2 weeks
- [ ] Performance acceptable (create lab < 10 minutes)
- [ ] Security reviewed
- [ ] User satisfaction positive

## Feature priority matrix

### 🔴 Critical (Blocks Terraform replacement)

- [x] Week 1: Route53 DNS provider
- [x] Week 1: Wildcard DNS records
- [x] Week 1: Cloud-init user data for disk partitioning
- [x] Week 2: Enhanced Ansible variable passing
- [x] Week 3: Deployment orchestration with SSH readiness
- [x] Week 5: TTL daemon (automated cleanup)
- [x] Week 5: Access information generation

### 🟡 Important (Full feature parity)

- [x] Week 2: Training lab mode support
- [x] Week 4: Cost tracking
- [x] Week 6: Multi-server coordination
- [x] Week 6: Budget management
- [x] Week 7-9: Web dashboard

### 🟢 Nice to have (Future enhancements)

- [ ] Future: AWS provider
- [ ] Future: DigitalOcean provider
- [ ] Future: VM snapshots
- [ ] Future: Private networks
- [ ] Future: Advanced approval workflows

## Technical decisions summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Database** | PostgreSQL 15+ | ACID, native locking, rich queries |
| **Web framework** | Echo | Better middleware, built-in validation |
| **Frontend** | React + Vite | Team familiarity (MinIO uses React) |
| **UI library** | Tailwind CSS + shadcn/ui | Modern, customizable |
| **Reverse proxy** | Caddy | Automatic HTTPS with Let's Encrypt |
| **DNS migration** | Keep Cloudflare, add Route53 | Gradual migration, support both |
| **Ansible reuse** | Embed existing roles | Mature, tested, significant investment |
| **State backend** | Local → PostgreSQL | Progressive enhancement |
| **Cloud-init** | Support both approaches | cloud-init + Ansible fallback |
| **Multi-cloud** | Hetzner first, AWS/DO later | Focus on primary use case |

## Critical path

These tasks are on the critical path (must be done in order):

1. Week 1: Database + storage abstraction → **Blocks:** All remote operations
2. Week 2: Web service foundation → **Blocks:** API client, dashboard
3. Week 3: CLI client updates → **Blocks:** Remote lab management
4. Week 7: Dashboard foundation → **Blocks:** All dashboard features

These can be done in parallel:
- Week 1: Route53 DNS (parallel with database)
- Week 4: Cost tracking (parallel with Week 5 daemon)
- Week 5: Access information (parallel with Week 6 budgets)

## Migration checklist for users

### Preparation
- [ ] Install storctl binary
- [ ] Run `storctl init`
- [ ] Configure providers in `~/.storctl/config.yaml`
- [ ] Add Route53 or Cloudflare credentials
- [ ] Test with small lab deployment

### Lab conversion
- [ ] Convert existing terraform.tfvars to lab.yaml
- [ ] Test lab creation: `storctl create lab test-01 -f lab.yaml`
- [ ] Verify DNS records created
- [ ] Verify Ansible runs successfully
- [ ] Check generated access scripts

### Team migration
- [ ] Train team on storctl commands
- [ ] Set up remote state (PostgreSQL)
- [ ] Migrate team members to remote mode
- [ ] Decommission Terraform state files
- [ ] Update documentation and runbooks

### Validation
- [ ] All team members can create/delete labs
- [ ] TTL enforcement working (labs auto-delete)
- [ ] Cost tracking accurate
- [ ] No runaway VMs for 1 month
- [ ] Team satisfied with workflow

## Command reference

### Terraform → storctl translation

| Terraform/Make | storctl equivalent |
|----------------|-------------------|
| `make setup` | `storctl create lab mylab -f lab.yaml` |
| `terraform output` | `storctl get lab mylab` |
| `make destroy` | `storctl delete lab mylab` |
| `make ansible-deploy` | `storctl install lab mylab` |
| `terraform workspace list` | `storctl get lab` |
| View access info | `storctl info lab mylab` |
| SSH to server | `storctl ssh lab mylab server-01` |
| Extend lifetime | `storctl extend lab mylab --ttl 3d` |
| Check costs | `storctl costs show mylab` |

### New commands (not in Terraform)

```bash
# Configure remote state
storctl config set-remote --url https://storctl.yourdomain.com --token TOKEN

# View all costs
storctl costs show [--owner X] [--project Y]

# Budget management
storctl budget create --owner pavel --limit 100
storctl budget list

# Audit logs
storctl audit show --since 7d

# Daemon status
storctl daemon status
```

## Infrastructure costs

### Development (local testing)
- PostgreSQL: Docker container (free)
- Total: **€0/month**

### Production (Hetzner)
- CX21 server (4GB RAM): €5.90/month
- PostgreSQL on same server: €0
- Backups (Object Storage): ~€1/month
- Domain + SSL: Free (Caddy + Let's Encrypt)
- Total: **~€7/month**

### Savings estimate
- One forgotten lab: €46.57/month
- Prevent 1 forgotten lab = ROI in 6 days
- Prevent 2-3 per quarter = ROI in 1 month

## Success metrics

### Technical metrics
- Lab creation time: < 10 minutes (from VM create to Ansible done)
- API response time: < 200ms (p95)
- Dashboard load time: < 2 seconds
- TTL enforcement accuracy: 100% (no missed expirations)
- Cost calculation accuracy: ±5% of actual bill

### User metrics
- Time to create lab: < 5 minutes (user effort)
- Errors requiring manual intervention: < 5%
- Team satisfaction: > 80% positive
- Documentation clarity: > 90% questions answered

### Business metrics
- Cost savings: > €100/month (prevented runaway VMs)
- Deployment frequency: > 20 labs/month
- Average lab lifetime: < 7 days (due to TTL)
- Budget alerts effectiveness: > 95% caught before exceeded

## Risk register

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Route53 integration issues | Medium | High | Test thoroughly, fallback to Cloudflare |
| Ansible role incompatibility | Low | High | Test with existing roles early |
| Database performance | Low | Medium | Connection pooling, indexes, monitoring |
| TTL daemon failures | Low | High | Monitoring alerts, manual cleanup backup |
| Migration disruption | Medium | Medium | Gradual rollout, keep Terraform accessible |
| Timeline slips | Medium | Medium | Prioritize critical features, defer nice-to-haves |
| Team capacity changes | Low | High | Good documentation, modular design |

## Next actions

### This week (Week 0)
1. **Review documents** with team and stakeholders
   - This roadmap
   - `REMOTE_STATE_IMPLEMENTATION_PLAN.md`
   - `TERRAFORM_MIGRATION_ANALYSIS.md`

2. **Get approvals**
   - Management approval for 10-12 week timeline
   - Budget approval for €7/month infrastructure
   - Resource allocation (developer time)

3. **Set up infrastructure**
   - Provision Hetzner CX21 server (or wait until Week 10)
   - Set up development environment (local PostgreSQL)
   - Create project board for task tracking

4. **Make key decisions**
   - DNS: Route53, Cloudflare, or both?
   - Domain: What domain for web service?
   - Backup: Retention policy (30, 90 days?)
   - Notifications: SMTP server details? Slack webhook?

### Next week (Week 1)
1. **Start implementation**
   - Database schema creation
   - Storage abstraction layer
   - Route53 DNS provider
   - Cloud-init templates

2. **Testing setup**
   - Hetzner test account
   - Route53 test zone
   - Continuous integration

## Resources

### Documentation
- Main plan: `docs/REMOTE_STATE_IMPLEMENTATION_PLAN.md`
- Migration analysis: `docs/TERRAFORM_MIGRATION_ANALYSIS.md`
- This roadmap: `docs/IMPLEMENTATION_ROADMAP.md`
- Architecture: `CLAUDE.md`

### Code references
- Current storctl: `cmd/`, `internal/`
- Terraform project: `../minio-lab-terraform/`
- Ansible roles: `../minio-lab-terraform/ansible/roles/`

### External resources
- Hetzner Cloud API: https://docs.hetzner.cloud/
- AWS Route53 SDK: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/route53
- Echo web framework: https://echo.labstack.com/
- React docs: https://react.dev/

## Questions and decisions log

### Open questions
1. Domain name for web service? (e.g., storctl.aistorlabs.com)
2. Initial admin user credentials?
3. Should we support both Route53 AND Cloudflare simultaneously?
4. DirectPV setup: reuse Ansible role or reimplement?
5. AWS/DigitalOcean providers: Priority or defer?

### Decisions made
- [2025-11-04] Use PostgreSQL for remote state (vs. S3)
- [2025-11-04] Build web service (vs. CLI-only)
- [2025-11-04] Timeline: 10-12 weeks
- [2025-11-04] Reuse Ansible roles (vs. rewrite)
- [2025-11-04] Hetzner first, AWS/DO later

## Version history

- **v1.0** (2025-11-04): Initial roadmap combining web service + Terraform migration
- **v1.1** (TBD): Updates after Week 1 implementation
- **v2.0** (TBD): Revised after team feedback

---

**This is a living document.** Update weekly with progress, learnings, and adjustments.

**Next review:** After Week 1 completion
