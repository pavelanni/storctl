# Migration from minio-lab-terraform to storctl

**Date:** 2025-11-04
**Status:** Analysis Complete
**Source Project:** ../minio-lab-terraform

## Executive summary

The minio-lab-terraform project is a **mature, feature-rich training lab deployment system** that combines Terraform (infrastructure) + Ansible (configuration) for multi-cloud MinIO training environments. It's more sophisticated than expected, with 15+ Terraform files, 10+ Ansible roles, and support for AWS, Hetzner Cloud, and DigitalOcean.

**Key finding:** This is NOT a simple "VM + DNS" setup. It includes:
- Multi-cloud provider abstraction
- Complex Ansible automation (Docker, K3s, DirectPV, HAProxy, VSCode, MinIO)
- Workspace-based isolation for concurrent deployments
- Training lab optimization mode (95% faster deployment)
- Template-based output generation (scripts, configs)

**Migration complexity:** Medium-High. storctl needs significant enhancements to match functionality.

## What the Terraform project does

### Infrastructure layer (Terraform)

1. **Multi-cloud VM provisioning**
   - Hetzner Cloud: CPX31 (4 vCPU, 8GB RAM), Ubuntu 24.04
   - AWS EC2: t3a.xlarge instances
   - DigitalOcean: s-4vcpu-8gb droplets
   - Conditional deployment (enable/disable per provider)

2. **DNS management**
   - Route53 or Cloudflare (configurable)
   - A records: `{server}.{domain}`
   - Wildcards: `*.{server}.{domain}`
   - TTL: 60 seconds

3. **SSH key management**
   - Auto-generates 4096-bit RSA key pairs per deployment
   - Uploads to all enabled cloud providers
   - Stores in `ansible/files/{deployment_name}/`

4. **Disk partitioning**
   - Via cloud-init user_data
   - Creates `/dev/sda2` (75% of disk) as XFS
   - Mounts at `/mnt/data`
   - NO separate volumes (just root disk partitioning)

5. **Workspace isolation**
   - Terraform workspaces per deployment
   - Independent state files
   - Multiple concurrent environments

### Configuration layer (Ansible)

**Main playbook** (`ansible/main.yml`) installs:

1. **Security** (minimal for training labs, full for production)
   - fail2ban, auditd, unattended-upgrades
   - Training mode: 30-45 seconds vs. production: 5-20 minutes

2. **Docker** (via geerlingguy.docker role)

3. **Lab server essentials**
   - 40+ packages (jq, yq, vim, maven, openjdk-21, etc.)
   - zsh with dotfiles
   - Docker Compose

4. **Certbot + HAProxy**
   - Let's Encrypt certificates
   - Reverse proxy with stats interface

5. **VSCode code-server**
   - Browser-based IDE (v4.90.3)
   - Password-protected

6. **K3s Kubernetes**
   - Lightweight single-node cluster

7. **DirectPV storage**
   - Loopback devices from sparse files
   - LVM: PV → VG → LV
   - Default: 4 x 10GB disks
   - Systemd service for persistence
   - Devices: `/dev/mapper/directpv-vg-lv0` through `lv3`

8. **MinIO deployment**
   - Types: single, distributed, native, core
   - Docker Compose or systemd
   - MC client setup
   - Optional license and config restoration

### Orchestration layer (Makefile)

40+ targets including:
- `make setup`: End-to-end deployment (Terraform → Ansible)
- `make destroy`: Teardown
- `make ansible-deploy`: Run Ansible only
- `make copy-license`: Copy license to servers
- `make repull-git`: Update lab repo

### Generated artifacts

Terraform generates 15+ files from templates:
- Ansible inventory
- ansible.cfg
- group_vars/all.yml
- Access scripts (SSH, MinIO UI)
- MC client setup scripts
- Site replication scripts
- Native MinIO startup scripts

All in `ansible/files/{deployment_name}/` and `ansible/inventory/{deployment_name}/`

## What storctl already has

### ✅ Already implemented

1. **Hetzner Cloud provider**
   - Create/delete servers
   - Server types and images
   - Location selection
   - Labels for organization

2. **Volume management**
   - Create/attach/delete volumes
   - Size and format configuration
   - Server attachment

3. **SSH key management**
   - Create key pairs
   - Upload to Hetzner
   - Key lifecycle

4. **DNS integration**
   - Cloudflare provider
   - A record creation/deletion
   - (Route53 NOT yet implemented)

5. **Ansible integration**
   - Generate inventory files
   - Run playbooks
   - Pass variables to Ansible

6. **Lab abstraction**
   - YAML-based lab definitions
   - Multi-server labs
   - Multi-volume labs
   - Labels and metadata

7. **TTL support**
   - Lab expiration timestamps
   - (Automated cleanup NOT yet implemented)

8. **Local state storage**
   - BoltDB database
   - Lab persistence
   - (Remote state NOT yet implemented)

9. **Lima provider**
   - Local VM testing
   - (For development, not production)

## What storctl needs to add

### 🔴 Critical gaps (must have for Terraform replacement)

1. **Route53 DNS provider**
   - Currently only Cloudflare
   - Need AWS Route53 support
   - Wildcard record support

2. **User data / cloud-init**
   - Disk partitioning on creation
   - Custom user data scripts
   - Template rendering

3. **Multi-server lab orchestration**
   - Currently creates servers, but no coordination
   - Need sequential creation for dependencies
   - Wait for server readiness before next steps

4. **Ansible variable passing**
   - Pass all lab spec variables to Ansible
   - Template rendering for group_vars
   - Support for 50+ configuration variables

5. **Deployment lifecycle orchestration**
   - Infrastructure phase (create VMs)
   - Wait phase (SSH readiness)
   - Configuration phase (run Ansible)
   - Status reporting throughout

6. **Access information generation**
   - SSH connection strings
   - Service URLs (MinIO, VSCode, HAProxy)
   - Credential summary
   - Copy-paste ready output

### 🟡 Important features (should have)

7. **AWS provider**
   - EC2 instance creation
   - Security group management
   - Key pair management
   - (Currently Hetzner only)

8. **DigitalOcean provider**
   - Droplet creation
   - SSH key management
   - (Nice to have for compatibility)

9. **Training lab optimization mode**
   - Pass flag to Ansible for fast mode
   - Skip slow security tasks
   - Document security tradeoffs

10. **DirectPV automation**
    - Create loopback devices
    - LVM setup
    - Systemd service deployment
    - (Or just document in Ansible role)

11. **Multi-cloud deployment**
    - Deploy to multiple clouds in single command
    - Provider selection logic
    - Conditional resource creation

12. **Workspace/environment isolation**
    - Multiple concurrent deployments
    - Name-based separation
    - Independent lifecycle

### 🟢 Nice to have (future enhancements)

13. **Template-based script generation**
    - MC client setup scripts
    - Site replication scripts
    - Access scripts

14. **License file management**
    - Copy license to servers
    - Store license in state

15. **Configuration restoration**
    - Restore MinIO cluster config from snapshot
    - Backup/restore workflows

16. **Auto-shutdown scheduling**
    - Time-based shutdown
    - Cost savings for idle labs

17. **Network management**
    - Private networks
    - Firewall rules
    - Security groups

18. **Snapshot support**
    - VM snapshots
    - Clone from snapshot
    - Backup workflows

## Feature comparison matrix

| Feature | minio-lab-terraform | storctl (current) | storctl (needed) |
|---------|---------------------|-------------------|------------------|
| **Hetzner Cloud VMs** | ✅ | ✅ | - |
| **AWS EC2** | ✅ | ❌ | 🔴 Optional |
| **DigitalOcean** | ✅ | ❌ | 🟢 Optional |
| **Route53 DNS** | ✅ | ❌ | 🔴 Required |
| **Cloudflare DNS** | ✅ | ✅ | - |
| **Wildcard DNS** | ✅ | ❌ | 🔴 Required |
| **SSH key generation** | ✅ | ✅ | - |
| **Cloud-init user data** | ✅ | ❌ | 🔴 Required |
| **Disk partitioning** | ✅ (cloud-init) | ❌ | 🔴 Required |
| **Separate volumes** | ❌ | ✅ | - (storctl better!) |
| **Ansible integration** | ✅ | ✅ (basic) | 🟡 Enhance |
| **Inventory generation** | ✅ | ✅ | 🟡 Enhance |
| **Variable passing** | ✅ (50+ vars) | ❌ | 🔴 Required |
| **Multi-cloud deployment** | ✅ | ❌ | 🟡 Optional |
| **Workspace isolation** | ✅ (Terraform) | ✅ (implicit) | - |
| **State management** | ✅ (per workspace) | ✅ (BoltDB) | 🟡 Enhance |
| **TTL/expiration** | ✅ (script) | ✅ (metadata) | 🔴 Automate |
| **Cost tracking** | ❌ | ❌ | 🟢 New feature |
| **Training mode** | ✅ | ❌ | 🟡 Pass to Ansible |
| **DirectPV setup** | ✅ (Ansible) | ❌ | 🟡 Ansible role |
| **Access scripts** | ✅ (generated) | ❌ | 🟡 Generate |
| **License management** | ✅ | ❌ | 🟢 Optional |

**Legend:**
- 🔴 Critical (blocks Terraform replacement)
- 🟡 Important (needed for full parity)
- 🟢 Nice to have (enhances experience)

## Migration strategy

### Phase 0: Ansible role reuse (No changes needed)

**Good news:** All Ansible roles from minio-lab-terraform can be reused as-is!

storctl already has Ansible integration. Just need to:
1. Copy roles from `../minio-lab-terraform/ansible/roles/` to storctl's embedded assets
2. Use same playbook structure
3. Pass required variables

**Action:** Minimal - just reference the existing Ansible repo or embed it.

### Phase 1: Critical infrastructure parity (Weeks 1-3)

**Goal:** storctl can replace Terraform for basic Hetzner + DNS deployment

**Tasks:**

1. **Route53 DNS provider** (Week 1)
   - Add AWS SDK for Go
   - Implement Route53 in `internal/dns/route53.go`
   - A records + wildcard support
   - Update DNS factory to support selection

2. **Cloud-init user data** (Week 1)
   - Add `UserData` field to ServerCreateOpts
   - Template for disk partitioning (from terraform user_data.tmpl)
   - Pass to Hetzner API

3. **Wildcard DNS records** (Week 1)
   - Modify DNS creation to add wildcard: `*.{server}.{domain}`
   - Both Route53 and Cloudflare

4. **Enhanced Ansible variable passing** (Week 2)
   - Generate group_vars/all.yml from lab spec
   - Support arbitrary variables in lab YAML
   - Template rendering

5. **Deployment orchestration improvements** (Week 2-3)
   - SSH readiness polling (replace fixed wait)
   - Sequential server creation for dependencies
   - Progress reporting
   - Error handling and rollback

6. **Access information output** (Week 3)
   - Generate SSH commands
   - Service URLs (with wildcard support)
   - Credential summary
   - Format options (text, JSON, shell script)

**Deliverable:** `storctl create lab mylab` can replace `make setup` for Hetzner + Route53/Cloudflare

### Phase 2: Feature parity (Weeks 4-6)

**Goal:** storctl matches all minio-lab-terraform features

**Tasks:**

1. **Training lab mode** (Week 4)
   - Add `trainingMode` field to lab spec
   - Pass to Ansible as variable
   - Document in lab templates

2. **Multi-server coordination** (Week 4)
   - Server roles (control-plane, worker)
   - Creation order enforcement
   - Dependency management

3. **Template-based outputs** (Week 5)
   - MC client setup scripts
   - Access scripts (SSH, services)
   - Site replication scripts
   - Store in `~/.storctl/labs/{lab_name}/`

4. **DirectPV automation** (Week 5)
   - Ansible role for loopback device setup
   - LVM configuration
   - Systemd service deployment
   - (Or just reuse existing role)

5. **Enhanced status and info** (Week 6)
   - `storctl info lab mylab`: Show all access details
   - `storctl ssh lab mylab [server-name]`: Direct SSH
   - `storctl logs lab mylab`: Show deployment logs

**Deliverable:** Feature parity with minio-lab-terraform for Hetzner deployments

### Phase 3: Multi-cloud support (Weeks 7-9, Optional)

**Goal:** Support AWS and DigitalOcean like Terraform version

**Tasks:**

1. **AWS EC2 provider** (Week 7-8)
   - Implement `internal/provider/aws/`
   - EC2 instance creation
   - Security group management
   - Key pair management
   - VPC networking

2. **DigitalOcean provider** (Week 8-9)
   - Implement `internal/provider/digitalocean/`
   - Droplet creation
   - SSH key management
   - Firewall rules

3. **Multi-cloud lab support** (Week 9)
   - Lab spec with multiple providers
   - Conditional resource creation
   - Cross-cloud operations

**Deliverable:** Can deploy to AWS, DigitalOcean, or mix of providers

### Phase 4: Enhancements beyond Terraform (Weeks 10-12)

**Goal:** Add features Terraform version lacks

Already covered in main implementation plan:
- Remote state storage (PostgreSQL)
- Web dashboard
- Cost tracking
- Automated TTL enforcement
- Budget alerts

## Updated lab YAML specification

Based on Terraform analysis, enhanced lab spec:

```yaml
apiVersion: v1
kind: Lab
metadata:
  name: training-lab-01
  labels:
    project: minio-training
    owner: pavel
    environment: development
    type: training

spec:
  ttl: 7d  # Auto-delete after 7 days

  # Provider selection (Hetzner primary, can add others)
  provider: hetzner
  location: nbg1

  # DNS configuration
  dns:
    provider: route53  # or cloudflare
    domain: lab.learn.min.io
    wildcard: true  # Create *.{server}.{domain} records

  # Server definitions
  servers:
    - name: server-01
      role: control-plane  # or worker
      type: cpx31  # Hetzner: cpx31, AWS: t3a.xlarge
      image: ubuntu-24.04
      userData: |  # Optional cloud-init script
        #cloud-config
        disk_setup:
          /dev/sda:
            table_type: gpt
            layout: true
            overwrite: false
        fs_setup:
          - device: /dev/sda2
            filesystem: xfs
        mounts:
          - ["/dev/sda2", "/mnt/data", "xfs", "defaults", "0", "2"]

    - name: server-02
      role: worker
      type: cpx31
      image: ubuntu-24.04

  # Volume definitions (optional, for separate disks)
  volumes:
    - name: data-01
      server: server-01
      size: 100  # GB
      format: xfs
      mountpoint: /mnt/data

  # Ansible configuration
  ansible:
    playbook: main.yml
    roles:
      - security
      - docker
      - lab_server
      - haproxy
      - code-server
      - k3s
      - directpv_setup
      - minio

    # Variables passed to Ansible
    vars:
      training_lab_mode: true  # Fast deployment mode
      install_docker: true
      install_k3s: true
      install_directpv: true
      directpv_disk_count: 4
      directpv_disk_size: 10  # GB
      minio_deployment_type: distributed
      minio_server_count: 4
      start_minio: true
      copy_license: true
      code_server_password: "{{ .Generated.CodeServerPassword }}"
      haproxy_stats_password: "{{ .Generated.HAProxyPassword }}"

  # Generated outputs (storctl creates these)
  outputs:
    - type: ssh_commands
      filename: ssh-access.sh
    - type: service_urls
      filename: service-urls.txt
    - type: mc_setup
      filename: mc-setup.sh
```

## Key design decisions

### 1. Ansible role reuse vs. reimplementation

**Decision:** Reuse existing Ansible roles from minio-lab-terraform

**Rationale:**
- Roles are mature and tested
- Represent significant investment (40+ hours development)
- No benefit to rewriting in Go
- Ansible provides flexibility for configuration changes

**Implementation:**
- Embed Ansible roles in storctl binary (like current templates)
- Or reference external Ansible repo
- storctl generates inventory and runs playbook

### 2. Cloud-init vs. separate disk partitioning

**Decision:** Support both approaches

**Rationale:**
- cloud-init: Fast, done at provision time (Terraform approach)
- Ansible: More flexible, can modify after creation
- Different clouds have different cloud-init support

**Implementation:**
- Add `userData` field to server spec (cloud-init)
- Keep Ansible-based disk setup as fallback
- User chooses approach

### 3. Multi-cloud abstraction depth

**Decision:** Deep abstraction with provider-specific extensions

**Rationale:**
- Common interface for: create, delete, status
- Provider-specific fields for advanced features
- Balance between simplicity and power

**Implementation:**
- Core interface: `CloudProvider`
- Provider-specific options: `HetznerOpts`, `AWSOpts`
- Smart defaults for common cases

### 4. Terraform workspace equivalent

**Decision:** Lab names are implicit workspaces

**Rationale:**
- Simpler mental model
- No separate workspace concept
- Lab name is unique identifier

**Implementation:**
- Each lab is independent
- State stored per lab in database
- `storctl list` shows all labs

### 5. State backend choice

**Decision:** Start with local BoltDB, add remote PostgreSQL

**Rationale:**
- Matches main implementation plan
- Progressive enhancement
- Backward compatibility

**Implementation:**
- Phase 1-3: Local BoltDB (for migration)
- Phase 4+: Remote PostgreSQL (for team collaboration)

## Migration workflow for users

### Step 1: Install storctl

```bash
# Download from releases or build from source
go build -o storctl .
mv storctl ~/.local/bin/

# Initialize config
storctl init
```

### Step 2: Configure providers and DNS

```bash
# Edit ~/.storctl/config.yaml
vi ~/.storctl/config.yaml

# Add Hetzner token, Route53 credentials, etc.
```

### Step 3: Convert Terraform lab to storctl YAML

```bash
# Helper tool (optional, can create manually)
storctl convert terraform.tfvars > lab.yaml

# Or create manually based on template
cp ~/.storctl/templates/lab.yaml my-lab.yaml
vi my-lab.yaml
```

### Step 4: Deploy lab

```bash
# Create lab (replaces: make setup)
storctl create lab my-lab --file lab.yaml

# Or use CLI flags (simple case)
storctl create lab my-lab \
  --provider hetzner \
  --servers 2 \
  --server-type cpx31 \
  --ttl 7d
```

### Step 5: Check status and access

```bash
# Show lab details (replaces: terraform output)
storctl get lab my-lab

# Show access information
storctl info lab my-lab

# SSH to server
storctl ssh lab my-lab server-01

# Extend TTL if needed
storctl extend lab my-lab --ttl 3d
```

### Step 6: Destroy lab

```bash
# Delete lab (replaces: make destroy)
storctl delete lab my-lab

# With confirmation
storctl delete lab my-lab --confirm
```

## Side-by-side command comparison

| Task | minio-lab-terraform | storctl |
|------|---------------------|---------|
| **Initialize** | Edit terraform.tfvars | `storctl init`, edit config.yaml |
| **Deploy** | `make setup` | `storctl create lab mylab -f lab.yaml` |
| **Check status** | `terraform output` | `storctl get lab mylab` |
| **Show access info** | `cat ansible/files/.../access.sh` | `storctl info lab mylab` |
| **SSH to server** | `ssh -i ansible/files/.../key.pem user@server` | `storctl ssh lab mylab server-01` |
| **Extend lifetime** | Edit tfvars, `make setup` | `storctl extend lab mylab --ttl 3d` |
| **List environments** | `terraform workspace list` | `storctl get lab` |
| **Destroy** | `make destroy` | `storctl delete lab mylab` |
| **Run Ansible only** | `make ansible-deploy` | `storctl install lab mylab` |
| **Check costs** | Manual Hetzner console | `storctl costs show mylab` |

## Risks and mitigations

### Risk 1: Ansible role compatibility

**Risk:** Existing Ansible roles may not work with storctl's inventory format

**Mitigation:**
- Test with existing roles before migration
- Ensure inventory format matches Terraform's output
- Maintain compatibility layer if needed

### Risk 2: Missing Terraform features

**Risk:** Users rely on Terraform features not in storctl

**Mitigation:**
- Comprehensive feature comparison (done above)
- Prioritize critical features in Phase 1
- Provide migration guide for edge cases

### Risk 3: Cloud-init differences across providers

**Risk:** cloud-init syntax varies between clouds

**Mitigation:**
- Provider-specific templates
- Validation before sending to API
- Fallback to Ansible-based setup

### Risk 4: State migration complexity

**Risk:** Migrating from Terraform state to storctl state

**Mitigation:**
- Don't migrate existing labs (let them expire)
- Fresh start with storctl
- Or: import tool to read Terraform state

### Risk 5: Multi-cloud testing burden

**Risk:** Need accounts on AWS, DO, Hetzner to test

**Mitigation:**
- Focus on Hetzner first (primary use case)
- Mock providers for testing
- Community testing for AWS/DO

## Recommendations

### Priority 1: Core infrastructure parity (Weeks 1-3)

**Must have** for Terraform replacement:
1. Route53 DNS provider
2. Wildcard DNS support
3. Cloud-init user data
4. Enhanced Ansible integration
5. Deployment orchestration

**Outcome:** Can replace Terraform for Hetzner + Route53/Cloudflare deployments

### Priority 2: Feature parity (Weeks 4-6)

**Should have** for comfortable migration:
1. Training lab mode
2. Multi-server coordination
3. Access information generation
4. DirectPV automation (or Ansible role reuse)

**Outcome:** Full feature parity with minio-lab-terraform

### Priority 3: Enhancements (Weeks 7-12)

**Nice to have** improvements over Terraform:
1. Remote state storage
2. Web dashboard
3. Cost tracking
4. Automated TTL cleanup
5. Multi-cloud support (AWS, DO)

**Outcome:** storctl is superior to Terraform approach

## Success metrics

### Migration complete when:

- ✅ Can deploy Hetzner VMs with Route53 DNS
- ✅ Wildcard DNS records work
- ✅ Disk partitioning via cloud-init works
- ✅ All Ansible roles execute successfully
- ✅ Generated access scripts match Terraform's
- ✅ Multi-server labs work
- ✅ TTL enforcement automated
- ✅ Team can create/destroy labs without Terraform knowledge

### Success indicators (1 month post-migration):

- Zero Terraform deployments (all via storctl)
- Reduced deployment time (automated TTL vs. manual)
- Improved team satisfaction (simpler workflow)
- Better cost control (TTL enforcement, cost tracking)

## Conclusion

The minio-lab-terraform project is **more complex than expected** but storctl can replace it with **6-12 weeks of focused development**. The key is prioritizing critical infrastructure parity first (Weeks 1-3), then adding features for full parity (Weeks 4-6), and finally enhancing beyond Terraform's capabilities (Weeks 7-12).

**Key advantages of migration:**
1. **Simpler:** One tool (storctl) vs. three (Terraform + Ansible + Makefile)
2. **Better state:** Centralized tracking vs. scattered workspace states
3. **Cost control:** Built-in TTL enforcement and cost tracking
4. **Better UX:** Intuitive commands vs. Makefile targets
5. **More features:** Web dashboard, budget alerts, etc.

**Key challenges:**
1. Route53 DNS implementation
2. Cloud-init user data support
3. Ansible integration enhancement
4. Multi-cloud abstraction (if AWS/DO needed)

**Recommendation:** Start with Priority 1 (Weeks 1-3) to prove feasibility, then proceed to full migration.

---

**Next steps:**
1. Review this analysis with team
1. Decide on timeline and priorities
1. Begin Phase 1: Route53 + cloud-init + Ansible enhancements
1. Test with real training lab deployment
1. Iterate based on findings
