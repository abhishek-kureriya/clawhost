variable "hetzner_api_token" {
  description = "Hetzner Cloud API token (read+write)"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Name prefix used for all Hetzner resources"
  type        = string
  default     = "clawhost"
}

variable "environment" {
  description = "Deployment environment label"
  type        = string
  default     = "production"
}

# ---- Server ----------------------------------------------------------------

variable "server_type" {
  description = <<-EOT
    Hetzner server type. Cost-effective options:
      cx11  → 1 vCPU, 2 GB RAM, 20 GB disk  (~€3.79/mo)  ← minimum
      cx21  → 2 vCPU, 4 GB RAM, 40 GB disk  (~€6.90/mo)  ← recommended
      cx31  → 2 vCPU, 8 GB RAM, 80 GB disk  (~€13.10/mo)
  EOT
  type        = string
  default     = "cx21"
}

variable "location" {
  description = "Hetzner datacenter location (nbg1, fsn1, hel1, ash, hil)"
  type        = string
  default     = "nbg1"
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
  description = "Go version to install on the server"
  type        = string
  default     = "1.23.4"
}

variable "node_version" {
  description = "Node.js major version to install (for npx MCP servers)"
  type        = string
  default     = "20"
}

variable "openclaw_repo_url" {
  description = "Git URL of the OpenClaw source repo to clone and build on the server. Leave empty to skip build (manual install required)."
  type        = string
  default     = ""
}

variable "repo_url" {
  description = "Git repository URL to clone on the server"
  type        = string
  default     = "https://github.com/abhishek-kureriya/clawhost.git"
}
