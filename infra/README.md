# ClawHost — Infrastructure

Terraform-based provisioning for ClawHost + OpenClaw on cloud servers.
Supports **DigitalOcean** (default) and **Hetzner Cloud** from a single shared bootstrap.

---

## How It Works — End to End

Running `make provision` from your local machine triggers this full flow:

```
Your Machine                 DigitalOcean / Hetzner          New Server
─────────────────            ──────────────────────          ──────────────────────────────
make do-login           →    Authenticates doctl CLI
make provision          →    Terraform creates:
                               • SSH key (from ~/.ssh/id_rsa.pub)
                               • Firewall (ports 22, 80, 443)
                               • Droplet (Ubuntu 24.04)
                               • Injects cloud-init.yaml script
                                                         ↓  First boot runs cloud-init:
                                                            1. apt update + upgrade
                                                            2. Install Docker + Compose
                                                            3. Install Node.js 20 (for npx MCP)
                                                            4. Install Go 1.23
                                                            5. git clone this repo
                                                            6. go build → clawhost-core binary
                                                            7. docker compose up →
                                                                 PostgreSQL
                                                                 OpenClaw UI  (:3000)
                                                                 Nginx        (:80)
                                                            8. systemd service for clawhost-core (:8080)
                                                            9. UFW firewall enabled
make logs-bootstrap     →                             ←  tail /var/log/clawhost-bootstrap.log
                                                            (watch progress live)
```

**After ~3 minutes you have:**

| What | Where |
|------|-------|
| OpenClaw AI UI | `http://<IP>:3000` |
| ClawHost Core API | `http://<IP>:8080` |
| PostgreSQL | internal, port 5432 |
| SSH access | `ssh root@<IP>` |

**Updating the server later** (`make deploy`) re-runs:
1. `git pull` — gets latest code
2. `go build` — rebuilds the binary
3. `systemctl restart clawhost-core` — applies changes
4. `docker compose pull && up -d` — updates containers

---

```
infra/
├── Makefile                            ← all commands (PROVIDER=do|hetzner)
├── .env.server.example                 ← copy to server as .env
├── terraform/
│   ├── cloud-init.yaml                 ← shared first-boot bootstrap script
│   ├── main.tf / variables.tf / …      ← Hetzner Cloud config
│   └── digitalocean/
│       ├── main.tf / variables.tf / …  ← DigitalOcean config
│       └── terraform.tfvars.example    ← copy → terraform.tfvars
├── docker/
│   └── openclaw.yml                    ← OpenClaw + PostgreSQL + Nginx stack
└── nginx/
    └── nginx.conf                      ← reverse proxy config
```

---

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) ≥ 1.6
- [doctl](https://docs.digitalocean.com/reference/doctl/how-to/install/) — DigitalOcean CLI
- A DigitalOcean or Hetzner account with an API token
- An SSH key pair (`~/.ssh/id_rsa` / `id_rsa.pub`)

---

## Step 0 — Login to DigitalOcean

Before provisioning, authenticate `doctl` with your account:

```bash
make -C infra do-login
# Prompts for your Personal Access Token
# Get one at: https://cloud.digitalocean.com/account/api/tokens
```

Or non-interactively with an env var:

```bash
DO_TOKEN=dop_v1_... make -C infra do-login
```

Verify the login and see existing resources:

```bash
make -C infra do-check
```

Example output (empty = no resources yet, that's fine):

```
==> Authenticated account
Email                    ...
==> Existing droplets
ID    Name    Status    Public IPv4    Region    Size Slug

==> SSH keys on account
ID    Name    FingerPrint
```

---

## DigitalOcean Droplet (default — from $6/mo)

### 1. Initialise

```bash
make -C infra init PROVIDER=do
# Creates infra/terraform/digitalocean/terraform.tfvars from example
```

### 2. Edit config

```bash
nano infra/terraform/digitalocean/terraform.tfvars
```

Key fields:

| Field | Description | Example |
|-------|-------------|---------|
| `do_token` | DigitalOcean personal access token | `"dop_v1_..."` |
| `region` | Datacenter region | `"ams3"` |
| `droplet_size` | Droplet slug | `"s-1vcpu-1gb"` ($6/mo) |
| `ssh_public_key_path` | Path to your public key | `"~/.ssh/id_rsa.pub"` |
| `allowed_ssh_ips` | IPs allowed to SSH | `["1.2.3.4/32"]` |

**Droplet size options:**

| Slug | vCPU | RAM | Cost |
|------|------|-----|------|
| `s-1vcpu-1gb` | 1 | 1 GB | $6/mo ← default |
| `s-1vcpu-2gb` | 1 | 2 GB | $12/mo |
| `s-2vcpu-4gb` | 2 | 4 GB | $24/mo |

**Available regions:** `ams3` · `nyc3` · `sgp1` · `lon1` · `fra1` · `blr1` · `tor1` · `syd1`

### 3. Preview

```bash
make -C infra plan PROVIDER=do
```

### 4. Provision

```bash
make -C infra provision PROVIDER=do
# ~30 sec to create the droplet, ~3 min for bootstrap to complete
```

### 5. Watch bootstrap

```bash
make -C infra logs-bootstrap PROVIDER=do
```

Once bootstrap completes:

| Service | URL |
|---------|-----|
| OpenClaw UI | `http://<SERVER_IP>:3000` |
| ClawHost Core API | `http://<SERVER_IP>:8080` |
| SSH | `ssh root@<SERVER_IP>` |

---

## Hetzner Cloud (from ~€3.79/mo)

```bash
make -C infra init PROVIDER=hetzner
# edit infra/terraform/terraform.tfvars — set hetzner_api_token

make -C infra plan PROVIDER=hetzner
make -C infra provision PROVIDER=hetzner
make -C infra logs-bootstrap PROVIDER=hetzner
```

**Server size options:**

| Type | vCPU | RAM | Cost |
|------|------|-----|------|
| `cx11` | 1 | 2 GB | ~€3.79/mo |
| `cx21` | 2 | 4 GB | ~€6.90/mo ← default |
| `cx31` | 2 | 8 GB | ~€13.10/mo |

---

## Day-2 Operations

```bash
# SSH into the server
make -C infra ssh

# Deploy latest code (git pull + rebuild + restart services)
make -C infra deploy

# Tail ClawHost Core API logs
make -C infra logs

# Tail OpenClaw container logs
make -C infra logs-openclaw

# Show service health summary
make -C infra status

# Issue a Let's Encrypt TLS cert (domain must point to server IP first)
make -C infra ssl DOMAIN=yourdomain.com

# Tear down all cloud resources
make -C infra destroy
```

---

## Server Environment

Copy `infra/.env.server.example` to `/opt/clawhost/app/.env` on the server and fill in your values before running services.

Key variables:

```bash
# Database
DB_PASSWORD=strong_password

# LLM
OPENAI_API_KEY=sk-...

# MCP Bridge services (optional — enable what you use)
GITHUB_PERSONAL_ACCESS_TOKEN=
SLACK_BOT_TOKEN=
TAVILY_API_KEY=

# Security
JWT_SECRET=64_char_random_string
WEBUI_SECRET_KEY=random_string
```

---

## What the Bootstrap Does

The `cloud-init.yaml` script runs automatically on first boot and:

1. Updates system packages
2. Installs **Docker** + Docker Compose
3. Installs **Node.js 20** (required for `npx` MCP servers)
4. Installs **Go 1.23** and builds `clawhost-core`
5. Clones this repository to `/opt/clawhost/app`
6. Starts **OpenClaw + PostgreSQL + Nginx** via Docker Compose
7. Registers `clawhost-core` as a **systemd service** (auto-restart on crash/reboot)
8. Configures **UFW firewall** (allow 22, 80, 443; deny everything else)
