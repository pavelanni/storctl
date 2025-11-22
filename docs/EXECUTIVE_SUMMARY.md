# storctl Enhancement - Executive Summary

**Date:** 2025-11-04
**Prepared for:** Management Review
**Timeline:** 8-10 weeks
**Cost:** €6.50/month infrastructure + development time

## Problem statement

Current Terraform-based lab management has critical issues:
- **State fragmentation** - Each team member has local Terraform state, causing coordination problems
- **Runaway VM costs** - Forgotten labs accumulate charges (€46.57/month per lab)
- **Manual cleanup** - No automated TTL enforcement
- **Complex workflow** - Requires Terraform + Ansible + Makefile knowledge
- **No cost visibility** - Can't track spend by owner/project
- **DNS orphans** - Manual VM deletion leaves DNS records intact

## Proposed solution

Enhance existing storctl tool with:
1. **Remote state storage** (PostgreSQL) - Team sees same state
1. **Web dashboard** - Cost visibility for management
1. **Automated TTL enforcement** - Labs auto-delete after expiration
1. **Cost tracking** - Real-time spend monitoring by owner/project
1. **Route53 DNS** - Replace Cloudflare, clean up records automatically
1. **Budget alerts** - Warnings before overspending

## Key findings from analysis

### What exists in storctl today
- ✅ Hetzner Cloud integration (mature)
- ✅ Basic Ansible integration
- ✅ Cloud-init support (just needs disk partitioning template)
- ✅ Volume management
- ✅ TTL metadata (not automated)
- ✅ Local state storage (BoltDB)

### What needs to be added
- 🔴 Route53 DNS provider (~3 days)
- 🔴 Disk partitioning cloud-init template (~2 days)
- 🔴 Remote state backend + web service (~2 weeks)
- 🔴 TTL automation daemon (~3 days)
- 🔴 Cost tracking (~3 days)
- 🔴 Web dashboard (~3 weeks)

### What we're NOT doing (scope reduced)
- ❌ AWS/DigitalOcean support (not used)
- ❌ Training lab optimization mode (no benefit)
- ❌ Multi-server coordination (not currently needed)

**Result:** Timeline reduced from 12 weeks to **8-10 weeks**

## Architecture

### Current (Terraform)
```
Developer Laptop → Terraform → Hetzner Cloud
                 ↓
              Terraform State (local)
                 ↓
              Ansible → Configure VMs
```

**Problems:** Fragmented state, no cost tracking, manual cleanup

### Proposed (storctl + web service)
```
Developer Laptop → storctl CLI → storctl Server (Hetzner CX21)
                                      ↓
                              PostgreSQL (state)
                                      ↓
                                 TTL Daemon
                                      ↓
                              Hetzner Cloud API
                                      ↓
                                 Route53 DNS
                                      ↓
                              Ansible (reused)
```

**Benefits:** Centralized state, automated cleanup, cost tracking, web dashboard

## Benefits

### Operational
- **Simplified workflow** - One tool instead of three (Terraform + Ansible + Makefile)
- **Automated cleanup** - TTL enforcement prevents forgotten labs (24/7 monitoring)
- **Better collaboration** - Shared state, no conflicts
- **Faster onboarding** - Intuitive CLI commands

### Financial
- **Cost control** - Budget alerts and spend tracking
- **Prevent waste** - Auto-delete expired labs
- **Visibility** - Real-time dashboard for management
- **ROI** - < 1 week if prevents one forgotten lab

### Compliance
- **Audit trail** - All operations logged
- **Cost attribution** - Track spend by owner/project
- **Lifecycle management** - Documented TTL policies

## Cost analysis

### Infrastructure costs
| Item | Monthly Cost |
|------|--------------|
| Hetzner CX21 server (4GB RAM) | €5.90 |
| Backups (Object Storage) | €0.50 |
| Domain + SSL | €0 (Let's Encrypt) |
| **Total** | **€6.50/month** |

### Development investment
- Timeline: 8-10 weeks
- Engineer time: 320-400 hours
- Opportunity cost: Deferred features elsewhere

### Return on investment

**Scenario analysis:**

| Forgotten Labs/Quarter | Monthly Waste | Annual Savings | ROI Period |
|------------------------|---------------|----------------|------------|
| 1 lab | €46.57 | €558.84 | 5 days |
| 2 labs | €93.14 | €1,117.68 | 3 days |
| 3 labs | €139.71 | €1,676.52 | 2 days |

**Conservative estimate:** Preventing 2-3 forgotten labs per quarter

**Expected annual savings:** €500-1,000

**Infrastructure cost:** €78/year

**Net savings:** €420-920/year

## Timeline

### Phase 1: Core functionality (Weeks 1-3)
- Database + Route53 + Cloud-init
- Web service + CLI remote mode
- **Deliverable:** Can create labs via API with Route53 DNS

### Phase 2: Cost control (Weeks 4-6)
- Cost tracking
- TTL automation daemon
- Budget management
- **Deliverable:** Labs auto-delete, costs visible

### Phase 3: Dashboard (Weeks 7-9)
- Web UI for management visibility
- Labs, costs, audit logs
- **Deliverable:** Management dashboard

### Phase 4: Production (Week 10)
- Deployment + documentation
- Team migration from Terraform
- **Deliverable:** Full replacement of Terraform workflow

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| **Route53 integration fails** | Keep Cloudflare as backup option |
| **Ansible incompatibility** | Reuse exact same roles from Terraform project |
| **Timeline slips** | Built-in buffer, prioritize MVP features first |
| **Team adoption resistance** | Gradual rollout, comprehensive training, CLI similar to kubectl |

## Success criteria

### Technical
- ✅ Zero runaway VMs for 30 days
- ✅ Cost tracking accurate (±5% of Hetzner bill)
- ✅ Lab creation < 8 minutes
- ✅ TTL enforcement 100% reliable

### User
- ✅ Team adoption: All members using storctl
- ✅ Satisfaction: > 85% positive feedback
- ✅ Learning curve: < 30 minutes for basic operations

### Business
- ✅ Cost savings: > €100/month
- ✅ Zero Terraform deployments for 2 weeks
- ✅ Management visibility: Weekly cost reports

## Recommendation

**Proceed with storctl enhancement project**

**Rationale:**
1. **Low risk** - Building on existing, functional codebase
1. **High ROI** - Payback in days, not months
1. **Reasonable timeline** - 8-10 weeks with reduced scope
1. **Strategic value** - Better control, visibility, compliance

**Next steps:**
1. Management approval for 8-10 week project
1. Answer open questions (domain, Route53 zone, SMTP server)
1. Begin Week 1: Database + Route53 implementation

## Alternatives considered

### Alternative 1: Keep Terraform + add remote state
**Cost:** Low (Terraform Cloud ~€200/month or self-hosted S3)
**Time:** 1-2 weeks
**Verdict:** ❌ Doesn't solve complexity, no cost tracking, no TTL automation

### Alternative 2: Switch to Pulumi
**Cost:** Medium (development time)
**Time:** 10-12 weeks (complete rewrite)
**Verdict:** ❌ Doesn't solve core problems, high risk

### Alternative 3: Build from scratch
**Cost:** High (6+ months)
**Time:** 6+ months
**Verdict:** ❌ Too expensive, too risky

### **Selected: Enhance storctl** ✅
**Cost:** €6.50/month + 8-10 weeks dev time
**Time:** 8-10 weeks (incremental enhancement)
**Verdict:** ✅ Best balance of risk, cost, timeline, and value

## Appendix: Command comparison

| Task | Terraform | storctl |
|------|-----------|---------|
| Deploy lab | `make setup` | `storctl create lab mylab -f lab.yaml` |
| Check status | `terraform output` | `storctl get lab mylab` |
| Access server | Copy SSH command | `storctl ssh lab mylab` |
| Extend lifetime | Edit tfvars, re-run | `storctl extend lab mylab --ttl 3d` |
| Check costs | Hetzner console | `storctl costs show` |
| Destroy lab | `make destroy`, type name | `storctl delete lab mylab` |

**Result:** Simpler, more intuitive workflow

---

**Prepared by:** Technical team
**Review date:** 2025-11-04
**Decision needed by:** [Management to specify]
**Questions:** See IMPLEMENTATION_ROADMAP_V2.md for technical details
