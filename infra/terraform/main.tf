terraform {
  required_version = ">= 1.6"
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.49"
    }
  }
}

provider "hcloud" {
  token = var.hetzner_api_token
}

# ---------------------------------------------------------------------------
# SSH Key
# ---------------------------------------------------------------------------
resource "hcloud_ssh_key" "clawhost" {
  name       = "${var.project_name}-key"
  public_key = file(var.ssh_public_key_path)
}

# ---------------------------------------------------------------------------
# Firewall
# ---------------------------------------------------------------------------
resource "hcloud_firewall" "clawhost" {
  name = "${var.project_name}-fw"

  # SSH
  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "22"
    source_ips = var.allowed_ssh_ips
  }

  # HTTP
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # HTTPS
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # ClawHost Core API (internal / dev access only)
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "8080"
    source_ips = var.allowed_ssh_ips
  }

  # OpenClaw UI
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "3000"
    source_ips = var.allowed_ssh_ips
  }

  # Allow all outbound
  rule {
    direction   = "out"
    protocol    = "tcp"
    port        = "any"
    destination_ips = ["0.0.0.0/0", "::/0"]
  }
  rule {
    direction   = "out"
    protocol    = "udp"
    port        = "any"
    destination_ips = ["0.0.0.0/0", "::/0"]
  }
}

# ---------------------------------------------------------------------------
# Cloud-init bootstrap
# ---------------------------------------------------------------------------
locals {
  user_data = templatefile("${path.module}/cloud-init.yaml", {
    repo_url     = var.repo_url
    project_name = var.project_name
    go_version   = var.go_version
    node_version = var.node_version
  })
}

# ---------------------------------------------------------------------------
# Server
# ---------------------------------------------------------------------------
resource "hcloud_server" "clawhost" {
  name        = var.project_name
  image       = "ubuntu-24.04"
  server_type = var.server_type   # cx21 = 2vCPU / 4GB / €6.90/mo
  location    = var.location      # nbg1, fsn1, hel1, ash, hil
  ssh_keys    = [hcloud_ssh_key.clawhost.id]
  user_data   = local.user_data

  firewall_ids = [hcloud_firewall.clawhost.id]

  labels = {
    project = var.project_name
    env     = var.environment
  }
}

# ---------------------------------------------------------------------------
# Optional: attach a floating IP for static addressing
# ---------------------------------------------------------------------------
# resource "hcloud_floating_ip" "clawhost" {
#   type          = "ipv4"
#   home_location = var.location
# }
# resource "hcloud_floating_ip_assignment" "clawhost" {
#   floating_ip_id = hcloud_floating_ip.clawhost.id
#   server_id      = hcloud_server.clawhost.id
# }
