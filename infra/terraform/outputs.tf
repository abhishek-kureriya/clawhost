output "server_ip" {
  description = "Public IPv4 address of the ClawHost server"
  value       = hcloud_server.clawhost.ipv4_address
}

output "server_ipv6" {
  description = "Public IPv6 address of the ClawHost server"
  value       = hcloud_server.clawhost.ipv6_address
}

output "server_id" {
  description = "Hetzner server ID"
  value       = hcloud_server.clawhost.id
}

output "ssh_command" {
  description = "Ready-to-use SSH command"
  value       = "ssh root@${hcloud_server.clawhost.ipv4_address}"
}

output "core_api_url" {
  description = "ClawHost Core API base URL"
  value       = "http://${hcloud_server.clawhost.ipv4_address}:8080"
}

output "openclaw_ui_url" {
  description = "OpenClaw web UI URL"
  value       = "http://${hcloud_server.clawhost.ipv4_address}:3000"
}
