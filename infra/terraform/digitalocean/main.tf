terraform {
  required_version = ">= 1.6"
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.40"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

# ---------------------------------------------------------------------------
# SSH Key
# ---------------------------------------------------------------------------
resource "digitalocean_ssh_key" "clawhost" {
  name       = "${var.project_name}-key"
  public_key = file(pathexpand(var.ssh_public_key_path))
}

# ---------------------------------------------------------------------------
# Firewall
# ---------------------------------------------------------------------------
resource "digitalocean_firewall" "clawhost" {
  name        = "${var.project_name}-fw"
  droplet_ids = [digitalocean_droplet.clawhost.id]

  # SSH
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = var.allowed_ssh_ips
  }

  # HTTP
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # HTTPS
  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # ClawHost Core API (restricted)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "8080"
    source_addresses = var.allowed_ssh_ips
  }

  # OpenClaw UI (restricted)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "3000"
    source_addresses = var.allowed_ssh_ips
  }

  # Allow all outbound TCP/UDP
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# ---------------------------------------------------------------------------
# Cloud-init bootstrap (shared with Hetzner config)
# ---------------------------------------------------------------------------
locals {
  user_data = templatefile("${path.module}/../cloud-init.yaml", {
    repo_url           = var.repo_url
    project_name       = var.project_name
    go_version         = var.go_version
    node_version       = var.node_version
    openclaw_repo_url  = var.openclaw_repo_url
  })
}

# ---------------------------------------------------------------------------
# Droplet
# ---------------------------------------------------------------------------
resource "digitalocean_droplet" "clawhost" {
  name     = var.project_name
  image    = "ubuntu-24-04-x64"
  size     = var.droplet_size   # s-1vcpu-2gb = $12/mo  ← default
  region   = var.region         # nyc3, ams3, sgp1, lon1, fra1, blr1 …
  ssh_keys = [digitalocean_ssh_key.clawhost.fingerprint]
  user_data = local.user_data

  monitoring = true
  backups    = var.enable_backups   # adds 20% to monthly cost

  tags = [var.project_name, var.environment]
}

# ---------------------------------------------------------------------------
# Optional: reserved IP for static addressing
# ---------------------------------------------------------------------------
# resource "digitalocean_reserved_ip" "clawhost" {
#   region = var.region
# }
# resource "digitalocean_reserved_ip_assignment" "clawhost" {
#   ip_address = digitalocean_reserved_ip.clawhost.ip_address
#   droplet_id = digitalocean_droplet.clawhost.id
# }
