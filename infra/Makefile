# ClawHost Infrastructure Makefile
# Usage: make <target>  (run from repo root or infra/)

SHELL := /bin/bash
TF_DIR := infra/terraform

# Auto-detect server IP from Terraform output (after provisioning)
SERVER_IP := $(shell cd $(TF_DIR) && terraform output -raw server_ip 2>/dev/null)

.PHONY: help init plan provision destroy ssh deploy logs status ssl clean-tf

# ─────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────
help:
	@echo ""
	@echo "  ClawHost Infrastructure Commands"
	@echo "  ─────────────────────────────────"
	@echo "  make init        Init Terraform providers"
	@echo "  make plan        Preview infrastructure changes"
	@echo "  make provision   Create/update Hetzner server"
	@echo "  make destroy     Destroy all cloud resources"
	@echo "  make ssh         SSH into the server"
	@echo "  make deploy      Push latest code & restart services"
	@echo "  make logs        Tail logs from the server"
	@echo "  make status      Show service health on the server"
	@echo "  make ssl         Issue Let's Encrypt cert (needs DOMAIN=)"
	@echo ""

# ─────────────────────────────────────────────
# Terraform
# ─────────────────────────────────────────────
init:
	@echo "==> Initialising Terraform..."
	@[ -f $(TF_DIR)/terraform.tfvars ] || { \
		cp $(TF_DIR)/terraform.tfvars.example $(TF_DIR)/terraform.tfvars; \
		echo "  ✏  Edit infra/terraform/terraform.tfvars before running 'make provision'"; \
	}
	cd $(TF_DIR) && terraform init

plan:
	cd $(TF_DIR) && terraform plan

provision:
	@echo "==> Provisioning Hetzner server..."
	cd $(TF_DIR) && terraform apply -auto-approve
	@echo ""
	@echo "  Server IP : $$(cd $(TF_DIR) && terraform output -raw server_ip)"
	@echo "  SSH       : $$(cd $(TF_DIR) && terraform output -raw ssh_command)"
	@echo ""
	@echo "  Bootstrap is running in the background (~3 min)."
	@echo "  Watch progress: make logs-bootstrap"

destroy:
	@echo "WARNING: This will delete all cloud resources."
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
