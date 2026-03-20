variable "do_token" {
  description = "DigitalOcean personal access token (read+write)"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Name prefix used for all DO resources"
  type        = string
  default     = "clawhost"
}

variable "environment" {
  description = "Deployment environment label"
  type        = string
  default     = "production"
}

# ---- Droplet ---------------------------------------------------------------

variable "droplet_size" {
  description = <<-EOT
    DigitalOcean droplet size slug. Cost-effective options:
      s-1vcpu-1gb    →  1 vCPU,  1 GB RAM  ($6/mo)   ← bare minimum
      s-1vcpu-2gb    →  1 vCPU,  2 GB RAM  ($12/mo)  ← recommended
      s-2vcpu-4gb    →  2 vCPU,  4 GB RAM  ($24/mo)
      s-4vcpu-8gb    →  4 vCPU,  8 GB RAM  ($48/mo)
  EOT
  type        = string
  default     = "s-1vcpu-2gb"
}

variable "region" {
  description = <<-EOT
    DigitalOcean region slug:
      nyc3 – New York 3       ams3 – Amsterdam 3
      sgp1 – Singapore 1      lon1 – London 1
      fra1 – Frankfurt 1      blr1 – Bangalore 1
      tor1 – Toronto 1        syd1 – Sydney 1
  EOT
  type        = string
  default     = "ams3"
}

variable "enable_backups" {
  description = "Enable weekly droplet backups (+20% monthly cost)"
  type        = bool
  default     = false
}

# ---- Access ---------------------------------------------------------------

variable "ssh_public_key_path" {
  description = "Path to your SSH public key file"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "allowed_ssh_ips" {
  description = "CIDR ranges allowed to SSH and reach dev ports (8080, 3000)"
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]  # restrict in production!
}

# ---- Software versions ----------------------------------------------------

variable "go_version" {
  description = "Go version to install on the droplet"
  type        = string
  default     = "1.23.4"
}

variable "node_version" {
  description = "Node.js major version to install (for npx MCP servers)"
  type        = string
  default     = "20"
}

variable "repo_url" {
  description = "Git repository URL to clone on the droplet"
  type        = string
  default     = "https://github.com/abhishek-kureriya/clawhost.git"
}
