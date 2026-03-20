# ClawHost Infrastructure Makefile
# Usage: make <target> [PROVIDER=do|hetzner]
#
#   PROVIDER=do       → DigitalOcean droplet  (default)
#   PROVIDER=hetzner  → Hetzner Cloud server

SHELL := /bin/bash

# Default cloud provider
PROVIDER ?= do

# Resolve Terraform directory based on PROVIDER
ifeq ($(PROVIDER),do)
  TF_DIR := infra/terraform/digitalocean
  PROVIDER_LABEL := DigitalOcean
else ifeq ($(PROVIDER),hetzner)
  TF_DIR := infra/terraform
  PROVIDER_LABEL := Hetzner Cloud
else
  $(error Unknown PROVIDER "$(PROVIDER)". Use: do | hetzner)
endif

# Auto-detect server IP from Terraform output (after provisioning)
SERVER_IP := $(shell cd $(TF_DIR) && terraform output -raw server_ip 2>/dev/null)

.PHONY: help do-login do-check init plan provision destroy ssh deploy logs logs-bootstrap logs-openclaw status ssl clean-tf

# ─────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────
help:
	@echo ""
	@echo "  ClawHost Infrastructure Commands"
	@echo "  ─────────────────────────────────"
	@echo "  PROVIDER=do (default) | PROVIDER=hetzner"
	@echo ""
	@echo "  make init                   Init Terraform providers"
	@echo "  make plan                   Preview infrastructure changes"
	@echo "  make provision              Create/update droplet ($(PROVIDER_LABEL))"
	@echo "  make destroy                Destroy all cloud resources"
	@echo "  make ssh                    SSH into the server"
	@echo "  make deploy                 Push latest code & restart services"
	@echo "  make logs                   Tail ClawHost Core API logs"
	@echo "  make logs-bootstrap         Tail first-boot bootstrap log"
	@echo "  make logs-openclaw          Tail OpenClaw container logs"
	@echo "  make status                 Show service health on the server"
	@echo "  make ssl DOMAIN=x.com       Issue Let's Encrypt cert"
	@echo ""
	@echo "  Examples:"
	@echo "    make provision                      # DigitalOcean (default)"
	@echo "    make provision PROVIDER=hetzner     # Hetzner Cloud"
	@echo ""

# ─────────────────────────────────────────────
# DigitalOcean CLI login
# ─────────────────────────────────────────────

# Authenticate doctl using an API token.
# If DO_TOKEN is set in env it uses that; otherwise prompts interactively.
do-login:
	@command -v doctl >/dev/null 2>&1 || { \
		echo "doctl not found. Install it first:"; \
		echo "  Linux : sudo snap install doctl"; \
		echo "  macOS : brew install doctl"; \
		echo "  Docs  : https://docs.digitalocean.com/reference/doctl/how-to/install/"; \
		exit 1; \
	}
	@if [ -n "$$DO_TOKEN" ]; then \
		echo "==> Authenticating doctl with DO_TOKEN env var..."; \
		doctl auth init --access-token "$$DO_TOKEN"; \
	else \
		echo "==> Authenticating doctl interactively..."; \
		echo "    Paste your DigitalOcean Personal Access Token when prompted."; \
		echo "    (Create one at https://cloud.digitalocean.com/account/api/tokens)"; \
		doctl auth init; \
	fi
	@echo ""
	@echo "Login successful. Run 'make do-check' to verify."

# Show the authenticated account and a summary of existing resources.
do-check:
	@command -v doctl >/dev/null 2>&1 || { echo "doctl not installed. Run 'make do-login' first."; exit 1; }
	@echo "==> Authenticated account"
	@doctl account get
	@echo ""
	@echo "==> Existing droplets"
	@doctl compute droplet list --format "ID,Name,Status,PublicIPv4,Region,SizeSlug"
	@echo ""
	@echo "==> SSH keys on account"
	@doctl compute ssh-key list --format "ID,Name,FingerPrint"

# ─────────────────────────────────────────────
# Terraform
# ─────────────────────────────────────────────
init:
	@echo "==> Initialising Terraform for $(PROVIDER_LABEL)..."
	@[ -f $(TF_DIR)/terraform.tfvars ] || { \
		cp $(TF_DIR)/terraform.tfvars.example $(TF_DIR)/terraform.tfvars; \
		echo "  ✏  Edit $(TF_DIR)/terraform.tfvars before running 'make provision'"; \
	}
	cd $(TF_DIR) && terraform init

plan:
	@echo "==> Planning for $(PROVIDER_LABEL)..."
	cd $(TF_DIR) && terraform plan

provision:
	@echo "==> Provisioning on $(PROVIDER_LABEL)..."
	cd $(TF_DIR) && terraform apply -auto-approve
	@echo ""
	@echo "  Server IP : $$(cd $(TF_DIR) && terraform output -raw server_ip)"
	@echo "  SSH       : $$(cd $(TF_DIR) && terraform output -raw ssh_command)"
	@echo "  Core API  : $$(cd $(TF_DIR) && terraform output -raw core_api_url)"
	@echo "  OpenClaw  : $$(cd $(TF_DIR) && terraform output -raw openclaw_ui_url)"
	@echo ""
	@echo "  Bootstrap is running in the background (~3 min)."
	@echo "  Watch progress: make logs-bootstrap"

destroy:
	@echo "WARNING: This will delete ALL $(PROVIDER_LABEL) resources."
	@read -p "Type 'yes' to confirm: " confirm; \
	[ "$$confirm" = "yes" ] && cd $(TF_DIR) && terraform destroy -auto-approve

clean-tf:
	rm -rf $(TF_DIR)/.terraform $(TF_DIR)/.terraform.lock.hcl

# ─────────────────────────────────────────────
# Server access
# ─────────────────────────────────────────────
ssh:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	ssh root@$(SERVER_IP)

logs:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	ssh root@$(SERVER_IP) "journalctl -u clawhost-core -f --no-pager"

logs-bootstrap:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	ssh root@$(SERVER_IP) "tail -f /var/log/clawhost-bootstrap.log"

logs-openclaw:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	ssh root@$(SERVER_IP) "docker compose -f /opt/clawhost/app/infra/docker/openclaw.yml logs -f openclaw"

status:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	@echo "==> Service status on $(SERVER_IP)"
	ssh root@$(SERVER_IP) " \
		echo '--- ClawHost Core API ---'; \
		systemctl status clawhost-core --no-pager -l; \
		echo ''; \
		echo '--- Docker containers ---'; \
		docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'; \
	"

# ─────────────────────────────────────────────
# Deployment
# ─────────────────────────────────────────────
deploy:
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found. Run 'make provision' first."; exit 1; }
	@echo "==> Deploying latest code to $(SERVER_IP)..."
	ssh root@$(SERVER_IP) " \
		cd /opt/clawhost/app && \
		git pull && \
		/usr/local/go/bin/go build -o /usr/local/bin/clawhost-core ./core/cmd && \
		systemctl restart clawhost-core && \
		docker compose -f infra/docker/openclaw.yml pull && \
		docker compose -f infra/docker/openclaw.yml up -d && \
		echo 'Deploy complete.' \
	"

# ─────────────────────────────────────────────
# SSL (optional, needs a domain pointed at server)
# ─────────────────────────────────────────────
ssl:
	@[ -n "$(DOMAIN)" ] || { echo "Usage: make ssl DOMAIN=yourdomain.com"; exit 1; }
	@[ -n "$(SERVER_IP)" ] || { echo "No server IP found."; exit 1; }
	@echo "==> Issuing Let's Encrypt cert for $(DOMAIN)..."
	ssh root@$(SERVER_IP) " \
		certbot --nginx -d $(DOMAIN) --non-interactive --agree-tos -m admin@$(DOMAIN) && \
		docker compose -f /opt/clawhost/app/infra/docker/openclaw.yml restart nginx \
	"
