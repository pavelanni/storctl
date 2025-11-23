# storctl Implementation Roadmap v2 (Revised)

**Date:** 2025-11-04 (Revised based on scope clarification)
**Timeline:** 8-10 weeks
**Goal:** Replace minio-lab-terraform with enhanced storctl + web service
**Branch:** `feature/remote-state-cost-tracking`

## Scope clarifications

Based on project review and current needs:

✅ **Focus on:**
- Hetzner Cloud only (no AWS/DigitalOcean)
- Route53 DNS migration (critical)
- Remote state + web service
- Cost tracking + TTL automation
- Reuse ALL Ansible from minio-lab-terraform

❌ **Out of scope:**
- Multi-cloud support (AWS, DigitalOcean) - not used
- Training lab optimization mode - not showing benefits
- Multi-server coordination - lower priority, not currently used

🔧 **Already exists:**
- Cloud-init UserData support (just needs disk partitioning template)
- Hetzner provider (mature)
- Ansible integration (basic)
- Volume management

## Revised timeline: 8-10 weeks

### Week 1: Foundation + Route53

**Goal:** Database setup + Route53 DNS provider

**Database tasks:**
- Create PostgreSQL schema and migrations
- Add storage abstraction (local BoltDB + PostgreSQL)
- Database connection management

**DNS tasks:**
- Implement Route53 DNS provider (`internal/dns/route53.go`)
- Add wildcard DNS support (both Route53 and Cloudflare)
- Update DNS factory for provider selection
- Configuration updates for Route53 credentials

**Deliverable:** Database ready + Route53 DNS working

**Testing checklist:**
- [ ] PostgreSQL migrations run successfully
- [ ] Can create A records in Route53
- [ ] Wildcard records (`*.server.domain`) work
- [ ] DNS cleanup on lab deletion works

### Week 2: Cloud-init disk partitioning + Web service foundation

**Goal:** Disk partitioning template + REST API basics

**Cloud-init tasks:**
- Create disk partitioning cloud-init template (from Terraform user_data.tmpl)
- Template system for merging user creation + disk setup
- Pass template to existing UserData field (already supported)
- Test partitioning on Hetzner VMs

**Web service tasks:**
- Create `cmd/storctl-server/` web service binary
- REST API core endpoints: POST/GET/DELETE /api/v1/labs
- Authentication middleware (API tokens)
- Database operations

**Deliverable:** VMs boot with partitioned disks + API accepts requests

**Testing checklist:**
- [ ] VM boots with `/mnt/data` mounted
- [ ] Disk partitioned correctly (sda1 25%, sda2 75%)
- [ ] API endpoints respond to requests
- [ ] Authentication works

### Week 3: CLI client + Ansible enhancements

**Goal:** Remote mode CLI + full Ansible variable passing

**CLI tasks:**
- API client package (`internal/client/`)
- Update CLI for remote/local mode switching
- `storctl config set-remote` command
- Backward compatibility

**Ansible tasks:**
- Enhanced variable passing (50+ vars from lab spec)
- Generate group_vars/all.yml from lab configuration
- SSH readiness polling before Ansible runs
- Deployment orchestration improvements

**Deliverable:** End-to-end lab creation via API with Ansible

**Testing checklist:**
- [ ] CLI switches between local and remote modes
- [ ] Lab creation via API works end-to-end
- [ ] Ansible receives all variables from lab spec
- [ ] SSH readiness detected before Ansible starts
- [ ] All Ansible roles from minio-lab-terraform execute successfully

### Week 4: Cost tracking + access information

**Goal:** Cost calculation and generated scripts

**Cost tasks:**
- Cost calculation engine (Hetzner pricing)
- Track costs over lab lifecycle
- API endpoints: GET /api/v1/costs
- CLI: `storctl costs show`
- Display cost in `storctl get lab`

**Access info tasks:**
- Generate SSH connection scripts
- Service URLs (MinIO, VSCode, HAProxy, K3s)
- MC client setup scripts
- Store in `~/.storctl/labs/{lab_name}/scripts/`
- `storctl info lab mylab` command

**Deliverable:** Cost visibility + easy server access

**Testing checklist:**
- [ ] Cost calculations accurate for Hetzner resources
- [ ] Costs tracked over time in database
- [ ] Generated scripts work (SSH, MC setup)
- [ ] Service URLs correct with wildcard DNS

### Week 5: TTL daemon + notifications

**Goal:** Automated cleanup

**Tasks:**
- Background TTL enforcement daemon (goroutine in server)
- Check expired labs every 15 minutes
- Email/Slack notification system
- Pre-deletion warnings (24h, 4h, 1h before)
- Safe cleanup with audit logging
- Daemon status endpoint

**Deliverable:** Labs auto-delete after TTL expiration

**Testing checklist:**
- [ ] Daemon detects expired labs
- [ ] Warnings sent at correct intervals
- [ ] Labs deleted automatically after expiration
- [ ] Notifications delivered (email/Slack)
- [ ] Cleanup is safe (no orphaned resources)

### Week 6: Budget management

**Goal:** Budget alerts and enforcement

**Tasks:**
- Budget CRUD API endpoints
- Alert generation when approaching thresholds
- CLI: `storctl budget list/create/delete`
- Track alert history
- Optional: budget enforcement (block creation if exceeded)

**Deliverable:** Budget control system

**Testing checklist:**
- [ ] Budgets created and stored
- [ ] Alerts trigger at correct percentages (80%, 90%, 100%)
- [ ] CLI budget commands work
- [ ] Alert history tracked (no spam)

### Week 7: Dashboard foundation

**Goal:** Web UI authentication and layout

**Tasks:**
- React + Vite frontend setup
- Authentication UI (login with API token)
- Layout and navigation structure
- Embed built assets in Go binary
- API integration client (Axios)

**Deliverable:** Can log into web dashboard

**Testing checklist:**
- [ ] Frontend builds and embeds in binary
- [ ] Login page works
- [ ] API token authentication successful
- [ ] Navigation between pages works

### Week 8: Core dashboard views

**Goal:** Labs list and cost dashboard

**Tasks:**
- Labs list view (table with filtering and sorting)
- Lab details page (servers, volumes, cost breakdown, timeline)
- Cost dashboard (total spend, projections, charts)
- Real-time updates via polling

**Deliverable:** Can view and manage labs via web

**Testing checklist:**
- [ ] Labs table shows all active labs
- [ ] Filters work (owner, project, status)
- [ ] Lab details page complete with all info
- [ ] Cost charts render correctly
- [ ] Data updates in real-time

### Week 9: Management views

**Goal:** Audit logs and budgets in UI

**Tasks:**
- Audit log viewer (filterable, searchable)
- Budget management UI
- Lab extend TTL form
- User/token management
- CSV export functionality

**Deliverable:** Complete web management interface

**Testing checklist:**
- [ ] Audit logs visible and filterable
- [ ] Budget CRUD works in UI
- [ ] TTL extension works
- [ ] Export to CSV successful

### Week 10: Deployment + documentation

**Goal:** Production deployment and docs

**Tasks:**
- Ansible playbook for server deployment
- Systemd service files
- Caddy configuration (automatic HTTPS)
- Backup scripts (PostgreSQL → Hetzner Object Storage)
- Installation guide
- User guide (CLI + web)
- API documentation
- Operations runbook

**Deliverable:** Production-ready system with docs

**Testing checklist:**
- [ ] Deployment automated with Ansible
- [ ] HTTPS working via Caddy
- [ ] Backups running successfully
- [ ] Documentation complete and clear
- [ ] End-to-end system testing passed

## Feature priorities (revised)

### 🔴 Critical (MVP)

- [x] Week 1: Route53 DNS provider with wildcard support
- [x] Week 1: PostgreSQL database and storage abstraction
- [x] Week 2: Disk partitioning cloud-init template
- [x] Week 2: Web service foundation
- [x] Week 3: CLI remote mode
- [x] Week 3: Ansible variable passing
- [x] Week 5: TTL daemon (automated cleanup)
- [x] Week 4: Access information generation

### 🟡 Important (Full experience)

- [x] Week 4: Cost tracking
- [x] Week 6: Budget management
- [x] Week 7-9: Web dashboard
- [x] Week 10: Production deployment

### 🟢 Future enhancements (Post-launch)

- [ ] Multi-server coordination (when needed)
- [ ] AWS provider (if requirements change)
- [ ] VM snapshots
- [ ] Private networks
- [ ] Approval workflows
- [ ] Advanced RBAC

## What simplified the scope

1. **Hetzner only** - Removed AWS EC2 and DigitalOcean providers
   - Saved: ~2 weeks of multi-cloud abstraction work
   - Simplified: Provider interface, testing, documentation

2. **Cloud-init already exists** - Just needs disk template
   - UserData field already in ServerCreateOpts
   - Hetzner provider already passes UserData to API
   - Server checker already waits for cloud-init completion
   - Just need to add disk_setup/fs_setup template
   - Saved: ~1 week

3. **Skip training mode** - Not showing benefits
   - Saved: ~2 days of implementation and testing
   - Simplified: Ansible role logic

4. **Multi-server lower priority** - Not currently used
   - Deferred: Server roles, creation ordering, dependencies
   - Can add later when needed
   - Saved: ~3 days

**Total time saved: ~3 weeks** → Timeline reduced from 12 weeks to 8-10 weeks

## Architecture overview

### Simplified architecture (Hetzner only)

```
┌──────────────────────────────────────────┐
│  Hetzner CX21 Server (~€6/month)         │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  storctl-server (Go)                │ │
│  │  - REST API                         │ │
│  │  - TTL daemon                       │ │
│  │  - Web UI (embedded)                │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  PostgreSQL 15                      │ │
│  │  - Labs, costs, budgets, audit     │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Caddy (reverse proxy)              │ │
│  │  - Automatic HTTPS                  │ │
│  └────────────────────────────────────┘ │
└──────────────────────────────────────────┘
            ↓
┌──────────────────────────────────────────┐
│  Hetzner Cloud API                        │
│  - VMs (CPX31 typically)                 │
│  - Volumes (100GB XFS)                   │
│  - SSH Keys                              │
│  - Cloud-init for disk partitioning      │
└──────────────────────────────────────────┘
            ↓
┌──────────────────────────────────────────┐
│  Route53 DNS                              │
│  - A records: server.domain              │
│  - Wildcards: *.server.domain            │
└──────────────────────────────────────────┘
            ↓
┌──────────────────────────────────────────┐
│  Ansible (from minio-lab-terraform)      │
│  - Roles: docker, k3s, directpv, etc.    │
│  - Playbook: main.yml                    │
└──────────────────────────────────────────┘
```

## Lab YAML spec (revised, simpler)

```yaml
apiVersion: v1
kind: Lab
metadata:
  name: training-lab-01
  labels:
    owner: pavel
    project: minio-training

spec:
  ttl: 7d

  # Hetzner Cloud configuration
  provider: hetzner
  location: nbg1

  # DNS configuration
  dns:
    provider: route53  # or cloudflare
    domain: lab.learn.min.io
    wildcard: true

  # Servers (typically 1 for current use case)
  servers:
    - name: server-01
      type: cpx31  # 4 vCPU, 8GB RAM
      image: ubuntu-24.04

  # Volumes (optional - separate disks beyond root)
  # Note: Root disk is automatically partitioned via cloud-init
  volumes: []  # Can add additional volumes if needed

  # Ansible configuration
  ansible:
    playbook: main.yml  # From minio-lab-terraform
    # Variables from lab spec passed automatically
    vars:
      install_docker: true
      install_k3s: true
      install_directpv: true
      directpv_disk_count: 4
      directpv_disk_size: 10
      minio_deployment_type: distributed
      code_server_password: "{{ .Generated.CodeServerPassword }}"
```

## Command reference (Terraform → storctl)

| Task | minio-lab-terraform | storctl |
|------|---------------------|---------|
| Deploy lab | `make setup` | `storctl create lab mylab -f lab.yaml` |
| Check status | `terraform output` | `storctl get lab mylab` |
| Access info | `cat ansible/files/.../access.sh` | `storctl info lab mylab` |
| SSH to server | `ssh -i key.pem user@server` | `storctl ssh lab mylab server-01` |
| Extend lifetime | Edit tfvars + `make setup` | `storctl extend lab mylab --ttl 3d` |
| List labs | `terraform workspace list` | `storctl get lab` |
| Destroy | `make destroy` | `storctl delete lab mylab` |
| Check costs | Hetzner console | `storctl costs show mylab` |
| View audit logs | N/A | `storctl audit show --since 7d` |

## Key technical decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Cloud providers** | Hetzner only | Only cloud in use, simplifies everything |
| **DNS** | Route53 + Cloudflare | Support both, gradual migration |
| **Database** | PostgreSQL 15+ | ACID, locking, queries |
| **Web framework** | Echo | Middleware, validation |
| **Frontend** | React + Vite | Team familiarity |
| **Reverse proxy** | Caddy | Automatic HTTPS |
| **Ansible** | Reuse from minio-lab-terraform | Mature, tested, saves weeks |
| **Cloud-init** | Enhance existing support | Infrastructure exists, just add template |
| **Multi-server** | Defer to Phase 2 | Not currently needed |

## Migration checklist

### Preparation
- [ ] Install storctl binary
- [ ] Configure `~/.storctl/config.yaml` (Hetzner token, Route53 credentials)
- [ ] Test basic lab creation locally

### Lab conversion
- [ ] Convert terraform.tfvars to lab.yaml format
- [ ] Test lab creation end-to-end
- [ ] Verify DNS records in Route53
- [ ] Verify Ansible roles execute correctly
- [ ] Check generated access scripts

### Remote state setup
- [ ] Deploy storctl-server to Hetzner
- [ ] Configure team members for remote mode
- [ ] Migrate to remote state storage
- [ ] Test concurrent lab operations

### Team migration
- [ ] Train team on storctl CLI
- [ ] Update runbooks and documentation
- [ ] Decommission Terraform workflows
- [ ] Monitor for 2 weeks

### Validation
- [ ] No runaway VMs for 30 days (TTL working)
- [ ] Cost tracking matches Hetzner bills
- [ ] Team satisfied with workflow
- [ ] All documentation complete

## Infrastructure costs

| Item | Cost |
|------|------|
| Hetzner CX21 server (4GB RAM) | €5.90/month |
| PostgreSQL (on same server) | €0 |
| Backups (Object Storage ~50GB) | ~€0.50/month |
| Domain + SSL | €0 (Let's Encrypt) |
| **Total** | **~€6.50/month** |

### ROI
- One forgotten lab (3 servers + 8 volumes): €46.57/month
- Prevention of 1 forgotten lab: ROI in 5 days
- Expected savings: €100-200/month

## Success metrics

### Technical
- Lab creation time: < 8 minutes (VM + Ansible)
- API response time: < 200ms (p95)
- TTL accuracy: 100% (no missed expirations)
- Cost accuracy: ±5% of Hetzner bill

### User
- Time to create lab: < 3 minutes (user effort)
- CLI learning time: < 30 minutes
- Team satisfaction: > 85% positive

### Business
- Cost savings: > €100/month
- Zero runaway VMs
- Lab average lifetime: < 7 days

## Risks and mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Route53 integration issues | Low | High | Test early, keep Cloudflare as backup |
| Ansible compatibility | Low | High | Use exact same roles, test Week 3 |
| Database performance | Low | Medium | Proper indexes, connection pooling |
| TTL daemon failure | Low | High | Monitoring alerts, manual cleanup backup |
| Timeline slips | Medium | Medium | Built-in buffer, prioritize MVP |

## Next steps

### This week
1. Review revised plan with team
1. Get management approval
1. Answer open questions:
   - Domain for web service?
   - Route53 hosted zone ID?
   - SMTP server for notifications?
1. Set up development PostgreSQL

### Week 1 starts when ready
1. Database schema implementation
1. Route53 DNS provider
1. Wildcard DNS support
1. Testing with real Hetzner VMs

## Open questions

1. **Domain name** for storctl web service? (e.g., storctl.aistorlabs.com)
1. **Route53 hosted zone** - which domain/zone?
1. **Email notifications** - SMTP server details?
1. **Slack notifications** - webhook URL?
1. **Initial admin** - who gets first API token?
1. **Backup retention** - 30 days? 90 days?

## Version history

- **v1.0** (2025-11-04): Initial combined roadmap
- **v2.0** (2025-11-04): Revised based on scope clarification
  - Removed AWS/DigitalOcean
  - Simplified cloud-init (already exists)
  - Removed training mode
  - Multi-server deferred
  - Timeline: 12 weeks → 8-10 weeks

---

**Next review:** After Week 1 completion
**Living document:** Update weekly with progress and learnings
