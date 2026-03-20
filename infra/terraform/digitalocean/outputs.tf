output "server_ip" {
  description = "Public IPv4 address of the ClawHost droplet"
  value       = digitalocean_droplet.clawhost.ipv4_address
}

output "server_ipv6" {
  description = "Public IPv6 address of the ClawHost droplet"
  value       = digitalocean_droplet.clawhost.ipv6_address
}

output "droplet_id" {
  description = "DigitalOcean droplet ID"
  value       = digitalocean_droplet.clawhost.id
}

output "droplet_urn" {
  description = "DigitalOcean droplet URN (useful for DO projects/billing)"
  value       = digitalocean_droplet.clawhost.urn
}

output "ssh_command" {
  description = "Ready-to-use SSH command"
  value       = "ssh root@${digitalocean_droplet.clawhost.ipv4_address}"
}

output "core_api_url" {
  description = "ClawHost Core API base URL"
  value       = "http://${digitalocean_droplet.clawhost.ipv4_address}:8080"
}

output "openclaw_ui_url" {
  description = "OpenClaw web UI URL"
  value       = "http://${digitalocean_droplet.clawhost.ipv4_address}:3000"
}

output "monthly_cost_usd" {
  description = "Estimated monthly cost for this droplet size"
  value       = "See https://www.digitalocean.com/pricing/droplets for ${var.droplet_size}"
}
