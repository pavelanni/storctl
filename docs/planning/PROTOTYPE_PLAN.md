# storctl Prototype Implementation Plan

**Goal:** Build working prototype in 4-6 weeks (part-time, side project)
**Approach:** Incremental - each week adds working functionality
**Demo target:** Show to manager with live TTL deletion

## Why prototype first?

- ✅ Proof over promises (working demo beats slides)
- ✅ Learn real problems early
- ✅ No approval needed upfront
- ✅ Better estimates after building
- ✅ Easier buy-in with working code

## Prototype scope

### Include (must have for demo)
- ✅ PostgreSQL storage (local dev)
- ✅ Route53 DNS provider
- ✅ Cloud-init disk partitioning
- ✅ Full Ansible integration
- ✅ Cost tracking
- ✅ TTL auto-deletion
- ✅ Bubble Tea TUI dashboard

### Exclude (add after approval)
- ❌ Web service + REST API
- ❌ Web dashboard (React)
- ❌ Budget management
- ❌ Production deployment
- ❌ Comprehensive documentation

## Week-by-week plan

### Week 1: PostgreSQL migration (8-12 hours)

**Goal:** Replace BoltDB with PostgreSQL

**Setup:**
```bash
# Install PostgreSQL locally
brew install postgresql@17
brew services start postgresql@17

# Create database
createdb storctl_dev

# Or use Docker
docker run -d \
  --name storctl-postgres \
  -e POSTGRES_PASSWORD=storctl \
  -e POSTGRES_DB=storctl_dev \
  -p 5432:5432 \
  postgres:17
```

**Schema:** Create `migrations/001_initial.sql`:
```sql
CREATE TABLE labs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    metadata JSONB NOT NULL,
    spec JSONB NOT NULL,
    status JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_labs_name ON labs(name);
CREATE INDEX idx_labs_deleted_at ON labs(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    details JSONB
);

CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
```

**Code structure:**
```
internal/
├── storage/
│   ├── interface.go        # Storage interface (new)
│   ├── local/
│   │   └── bbolt.go        # BoltDB impl (move from lab.go)
│   └── postgres/
│       └── postgres.go     # PostgreSQL impl (new)
```

**Tasks:**
1. Create storage interface:
```go
// internal/storage/interface.go
package storage

import "github.com/pavelanni/storctl/internal/types"

type Storage interface {
    Save(lab *types.Lab) error
    Get(name string) (*types.Lab, error)
    List() ([]*types.Lab, error)
    Delete(name string) error
    Close() error
}
```

1. Implement PostgreSQL storage:
```go
// internal/storage/postgres/postgres.go
package postgres

import (
    "database/sql"
    "encoding/json"
    "fmt"

    _ "github.com/lib/pq"
    "github.com/paveni/storctl/internal/config"
    "github.com/pavelanni/storctl/internal/types"
)

type Storage struct {
    db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
    connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s",
        cfg.Storage.Postgres.Host,
        cfg.Storage.Postgres.Port,
        cfg.Storage.Postgres.Database,
        cfg.Storage.Postgres.User,
        cfg.Storage.Postgres.Password)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres db: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping postgres db: %w", err)
    }

    return &Storage{db: db}, nil
}

func (s *Storage) Save(lab *types.Lab) error {
    metadata, _ := json.Marshal(lab.ObjectMeta)
    spec, _ := json.Marshal(lab.Spec)
    status, _ := json.Marshal(lab.Status)

    _, err := s.db.Exec(`
        INSERT INTO labs (name, metadata, spec, status)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (name) DO UPDATE
        SET metadata = $2, spec = $3, status = $4, updated_at = NOW()
    `, lab.Name, metadata, spec, status)

    return err
}

func (s *Storage) Get(name string) (*types.Lab, error) {
    var metadata, spec, status []byte

    err := s.db.QueryRow(`
        SELECT metadata, spec, status
        FROM labs
        WHERE name = $1 AND deleted_at IS NULL
    `, name).Scan(&metadata, &spec, &status)

    if err != nil {
        return nil, err
    }

    lab := &types.Lab{}
    json.Unmarshal(metadata, &lab.ObjectMeta)
    json.Unmarshal(spec, &lab.Spec)
    json.Unmarshal(status, &lab.Status)
    lab.Name = name

    return lab, nil
}

// ... List, Delete, Close methods
```

1. Update config to support both:
```yaml
# config.yaml
storage:
  type: postgres  # or "local" for boltdb
  postgres:
    host: localhost
    port: 5432
    database: storctl_dev
    user: postgres
    password: storctl
```

1. Update lab manager to use storage interface

**Testing:**
```bash
# Run migrations
psql storctl_dev < migrations/001_initial.sql

# Test lab creation
storctl create lab test-01 -f examples/lab.yaml

# Verify in database
psql storctl_dev -c "SELECT name, created_at FROM labs;"

# Test retrieval
storctl get lab test-01

# Test deletion
storctl delete lab test-01
```

**Success criteria:**
- [ ] All existing commands work with PostgreSQL
- [ ] Can switch between BoltDB and PostgreSQL via config
- [ ] Data persists across restarts
- [ ] No regressions in functionality

### Week 2: Route53 DNS provider (6-10 hours)

**Goal:** Create/delete Route53 records

**Dependencies:**
```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/route53
go get github.com/aws/aws-sdk-go-v2/credentials
```

**Your OTP credential flow:**
```bash
# Your shell command (example)
$ get-aws-creds --otp $(op item get AWS --otp)
export AWS_ACCESS_KEY_ID=ASIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...

# Or output JSON
$ get-aws-creds --json
{"AccessKeyId": "ASIA...", "SecretAccessKey": "...", "SessionToken": "..."}
```

**Code structure:**
```go
// internal/dns/route53.go
package dns

import (
    "context"
    "os/exec"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/route53"
    "github.com/aws/aws-sdk-go-v2/service/route53/types"
    "github.com/aws/aws-sdk-go-v2/credentials"
)

type Route53Provider struct {
    client   *route53.Client
    hostedZoneID string
    domain   string
}

func NewRoute53(hostedZoneID, domain, credCmd string) (*Route53Provider, error) {
    // Run your OTP credential command
    creds, err := getOTPCredentials(credCmd)
    if err != nil {
        return nil, err
    }

    // Create AWS config with temporary credentials
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithCredentialsProvider(
            credentials.NewStaticCredentialsProvider(
                creds.AccessKeyID,
                creds.SecretAccessKey,
                creds.SessionToken,
            ),
        ),
    )
    if err != nil {
        return nil, err
    }

    client := route53.NewFromConfig(cfg)

    return &Route53Provider{
        client:       client,
        hostedZoneID: hostedZoneID,
        domain:       domain,
    }, nil
}

func (r *Route53Provider) CreateARecord(name, ip string) error {
    fqdn := fmt.Sprintf("%s.%s", name, r.domain)

    _, err := r.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
        HostedZoneId: &r.hostedZoneID,
        ChangeBatch: &types.ChangeBatch{
            Changes: []types.Change{
                {
                    Action: types.ChangeActionCreate,
                    ResourceRecordSet: &types.ResourceRecordSet{
                        Name: &fqdn,
                        Type: types.RRTypeA,
                        TTL:  aws.Int64(60),
                        ResourceRecords: []types.ResourceRecord{
                            {Value: &ip},
                        },
                    },
                },
            },
        },
    })

    return err
}

func (r *Route53Provider) CreateWildcard(name, ip string) error {
    wildcard := fmt.Sprintf("*.%s.%s", name, r.domain)
    // Similar to CreateARecord but with wildcard name
}

func (r *Route53Provider) DeleteRecord(name string) error {
    // List existing records first, then delete
}

func getOTPCredentials(cmd string) (*AWSCredentials, error) {
    out, err := exec.Command("bash", "-c", cmd).Output()
    if err != nil {
        return nil, err
    }

    var creds AWSCredentials
    json.Unmarshal(out, &creds)
    return &creds, nil
}

type AWSCredentials struct {
    AccessKeyID     string `json:"AccessKeyId"`
    SecretAccessKey string `json:"SecretAccessKey"`
    SessionToken    string `json:"SessionToken"`
}
```

**Config update:**
```yaml
dns:
  provider: route53
  route53:
    hosted_zone_id: Z1234567890ABC
    domain: lab.learn.min.io
    credential_command: "get-aws-creds --json"  # Your OTP script
```

**Testing:**
```bash
# Test record creation
storctl create lab test-dns -f examples/lab.yaml

# Verify in Route53 console or CLI
aws route53 list-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --query "ResourceRecordSets[?Name=='test-dns-server-01.lab.learn.min.io.']"

# Verify wildcard
dig *.test-dns-server-01.lab.learn.min.io
dig anything.test-dns-server-01.lab.learn.min.io

# Test deletion
storctl delete lab test-dns

# Verify records gone
```

**Success criteria:**
- [ ] A record created in Route53
- [ ] Wildcard record created
- [ ] Records deleted on lab deletion
- [ ] OTP credentials work
- [ ] Falls back gracefully if credentials expired

### Week 3: Cloud-init disk partitioning (4-6 hours)

**Goal:** VMs boot with `/mnt/data` partition

**Template:**
```yaml
# internal/config/constants.go - add disk partitioning
DefaultCloudInitWithDisk = `#cloud-config
users:
- name: ansible
  gecos: Ansible User
  groups: users,admin,wheel,sudo
  sudo: ALL=(ALL) NOPASSWD:ALL
  shell: /bin/bash
  ssh_authorized_keys:
  - %s

disk_setup:
  /dev/sda:
    table_type: gpt
    layout: true
    overwrite: false

fs_setup:
  - device: /dev/sda2
    filesystem: xfs
    partition: auto

mounts:
  - ["/dev/sda2", "/mnt/data", "xfs", "defaults,nofail", "0", "2"]

growpart:
  mode: auto
  devices: ['/dev/sda1']
  ignore_growroot_disabled: false

package_update: true
package_upgrade: true

power_state:
  mode: reboot
  message: Rebooting after setup
  condition: test -f /var/run/reboot-required
`
```

**Update server creation:**
```go
// internal/provider/hetzner/server.go
userData := fmt.Sprintf(config.DefaultCloudInitWithDisk, publicKey)

serverOpts := hcloud.ServerCreateOpts{
    // ...
    UserData: userData,
}
```

**Testing:**
```bash
# Create lab
storctl create lab test-disk -f examples/lab.yaml

# Wait for creation
# SSH to server
storctl ssh lab test-disk server-01

# Verify partition
df -h | grep /mnt/data
# Should show: /dev/sda2  ... /mnt/data

# Verify size (should be ~75% of disk)
lsblk
```

**Success criteria:**
- [ ] `/mnt/data` mounted on boot
- [ ] XFS filesystem
- [ ] Size is ~75% of root disk
- [ ] Survives reboot

### Week 4: Ansible integration (8-12 hours)

**Goal:** Full Ansible deployment works

**Tasks:**
1. Copy Ansible roles from minio-lab-terraform:
```bash
# Option 1: Embed in storctl
cp -r ../minio-lab-terraform/ansible/roles/ assets/ansible/roles/
cp ../minio-lab-terraform/ansible/main.yml assets/ansible/playbooks/

# Option 2: Reference external (simpler for prototype)
# Just point to ../minio-lab-terraform/ansible/
```

1. Enhanced inventory generation:
```go
// internal/lab/ansible.go - enhance CreateAnsibleInventoryFile

// Add ALL variables from lab spec
allVars := map[string]any{
    "ansible_user": ansibleUser,
    "ansible_ssh_private_key_file": sshKeyPath,

    // From lab.Spec.Ansible.Vars
    "install_docker": lab.Spec.Ansible.Vars["install_docker"],
    "install_k3s": lab.Spec.Ansible.Vars["install_k3s"],
    "directpv_disk_count": lab.Spec.Ansible.Vars["directpv_disk_count"],
    // ... all other vars
}

// Write group_vars/all.yml
groupVarsPath := filepath.Join(ansibleDir, "group_vars", "all.yml")
groupVarsData, _ := yaml.Marshal(allVars)
os.WriteFile(groupVarsPath, groupVarsData, 0644)
```

1. SSH readiness check:
```go
// Before running Ansible, wait for SSH
for _, server := range lab.Status.Servers {
    if err := waitForSSH(server.Status.PublicNet.IPv4.IP, sshKeyPath, 5*time.Minute); err != nil {
        return fmt.Errorf("server %s not ready: %w", server.Name, err)
    }
}
```

**Testing:**
```bash
# Full deployment
storctl create lab test-full -f examples/lab.yaml

# Should see:
# ✓ Creating VM...
# ✓ Creating DNS...
# ✓ Waiting for SSH...
# ✓ Running Ansible...
#   - security role
#   - docker role
#   - k3s role
#   - directpv_setup role
#   - minio role
# ✓ Lab ready!

# SSH and verify
storctl ssh lab test-full server-01

# Check installations
docker --version
kubectl get nodes
kubectl directpv info
```

**Success criteria:**
- [ ] Ansible runs without errors
- [ ] Docker installed and running
- [ ] K3s cluster ready
- [ ] DirectPV disks discovered
- [ ] MinIO deployed (if specified)

### Week 5: Cost tracking (4-6 hours)

**Goal:** See costs for every lab

**Code:**
```go
// internal/cost/pricing.go
package cost

var HetznerPricing = map[string]float64{
    // Servers (€/month)
    "cx11": 3.79,
    "cx21": 5.90,
    "cx22": 6.90,
    "cpx11": 4.90,
    "cpx21": 8.90,
    "cpx31": 13.90,
    // ... add all types
}

var HetznerVolumePricePerGB = 0.044 // €/GB/month

// internal/cost/calculator.go
func CalculateLabCost(lab *types.Lab) float64 {
    var totalHourly float64

    // Server costs
    for _, server := range lab.Status.Servers {
        monthlyPrice := HetznerPricing[server.Spec.ServerType]
        hourlyPrice := monthlyPrice / 730 // hours per month
        totalHourly += hourlyPrice
    }

    // Volume costs
    for _, volume := range lab.Status.Volumes {
        monthlyPrice := float64(volume.Spec.Size) * HetznerVolumePricePerGB
        hourlyPrice := monthlyPrice / 730
        totalHourly += hourlyPrice
    }

    // Calculate accumulated cost
    hoursRunning := time.Since(lab.Status.Created).Hours()
    return totalHourly * hoursRunning
}

func ProjectMonthlyCost(lab *types.Lab) float64 {
    hourlyRate := CalculateLabCost(lab) / time.Since(lab.Status.Created).Hours()
    return hourlyRate * 730
}
```

**CLI commands:**
```go
// cmd/costs.go
func (c *costsCmd) run(cmd *cobra.Command, args []string) error {
    labs, _ := labManager.List()

    var total float64
    for _, lab := range labs {
        cost := cost.CalculateLabCost(lab)
        total += cost
        fmt.Printf("%-20s  €%.2f\n", lab.Name, cost)
    }

    fmt.Printf("\nTotal: €%.2f\n", total)
    fmt.Printf("Monthly projection: €%.2f\n", projectMonthly(labs))
}
```

**Update lab display:**
```go
// cmd/get_lab.go - add cost to output
cost := cost.CalculateLabCost(lab)
projection := cost.ProjectMonthlyCost(lab)

fmt.Printf("Current cost: €%.2f\n", cost)
fmt.Printf("Monthly projection: €%.2f\n", projection)
```

**Success criteria:**
- [ ] Cost shown for each lab
- [ ] Cost increases over time
- [ ] Monthly projection calculated
- [ ] Costs by owner/project

### Week 6: TTL daemon + Bubble Tea TUI (8-12 hours)

**Goal:** Auto-delete + visual dashboard

**TTL Daemon:**
```go
// cmd/daemon.go (or integrate into server later)
func StartTTLDaemon(labManager *lab.ManagerSvc) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        labs, _ := labManager.List()

        for _, lab := range labs {
            if lab.Status.DeleteAfter.Before(time.Now()) {
                log.Printf("Deleting expired lab: %s", lab.Name)

                // Send notification
                sendNotification(lab)

                // Delete lab
                labManager.Delete(lab.Name, false)
            }
        }
    }
}
```

**Bubble Tea TUI:**
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
```

```go
// cmd/tui.go
package main

import (
    "fmt"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type model struct {
    labs []types.Lab
    costs map[string]float64
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        fetchLabs,
        tea.Tick(time.Second, func(t time.Time) tea.Msg {
            return tickMsg(t)
        }),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    case tickMsg:
        return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
            return tickMsg(t)
        })
    }
    return m, nil
}

func (m model) View() string {
    s := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("42")).
        Render("storctl Dashboard")

    s += "\n\n"

    // Lab list
    for _, lab := range m.labs {
        ttl := time.Until(lab.Status.DeleteAfter)
        cost := m.costs[lab.Name]

        s += fmt.Sprintf("%-20s  Status: %-10s  Cost: €%.2f  TTL: %s\n",
            lab.Name,
            lab.Status.State,
            cost,
            ttl.Round(time.Minute),
        )
    }

    s += "\n\nPress q to quit"
    return s
}

func main() {
    p := tea.NewProgram(initialModel())
    p.Run()
}
```

**Run TUI:**
```bash
storctl tui
```

**Demo value:**
- Real-time lab list
- TTL countdowns (updating every second!)
- Cost tracking
- Status changes visible live

**Success criteria:**
- [ ] TTL daemon detects expired labs
- [ ] Labs auto-deleted
- [ ] TUI shows live lab status
- [ ] TUI shows costs
- [ ] TUI updates in real-time

## Demo script (15 minutes)

```bash
# 1. Show TUI dashboard
$ storctl tui
[Shows empty dashboard]

# 2. Create lab (in another terminal)
$ storctl create lab demo-01 -f lab.yaml --ttl 10m
Creating lab demo-01...
✓ VM created
✓ DNS created
✓ Ansible complete
✓ Lab ready! Expires in 10 minutes

# 3. Watch TUI update (auto-refreshes)
[Shows demo-01 appear, cost accumulating, TTL counting down]

# 4. Show costs
$ storctl costs show
Lab        Current Cost  Monthly Projection
demo-01    €0.03         €6.90

# 5. SSH to server
$ storctl ssh lab demo-01
[Show K3s, DirectPV, MinIO running]

# 6. Watch TTL expire (wait 10 min or adjust for demo)
[TUI shows TTL reaching 0, lab disappears]

# 7. Confirm deletion
$ storctl get lab demo-01
Error: lab demo-01 not found
```

## Success metrics for prototype

### Technical
- ✅ Labs in PostgreSQL (not BoltDB)
- ✅ Route53 DNS working
- ✅ Disk partitioned automatically
- ✅ Ansible deploys successfully
- ✅ TTL daemon deletes expired labs
- ✅ TUI shows real-time status

### Demo impact
- ✅ Simpler than Terraform (one command)
- ✅ Cost visibility (new capability)
- ✅ TTL automation (solves main problem)
- ✅ Cool visual (Bubble Tea impresses)

## After successful demo

### If manager approves
1. Clean up prototype code
1. Add proper error handling
1. Write tests
1. Add web service + API
1. Build React dashboard (optional)
1. Production deployment

### If needs more proof
1. Use for your next course
1. Get team feedback
1. Iterate and demo again

## Getting started

### This weekend (4-6 hours)
1. Set up PostgreSQL locally
1. Create migrations
1. Implement Storage interface
1. Test basic CRUD with PostgreSQL

### Week 1 goal
- All existing commands work with PostgreSQL
- Ready to add Route53

---

**Questions?** Start with PostgreSQL migration - it's the foundation for everything else.
