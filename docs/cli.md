# Open Source OpenClaw CLI (Local / Self-Managed)

This guide is for the **open source core only** and focuses on OpenClaw self-hosting commands.

It does **not** require the commercial `hosting-service/` code and does **not** mix in cloud provisioning steps.

## What You Get in OSS Mode

- Run the core API on your own machine
- Use the basic local dashboard
- Manage local CLI state (profiles, backup, restore)

## Quick Start (Local Machine)

```bash
# 1) Build CLI
make build

# 2) Run core API + basic dashboard
./clawhost

# 3) Check health
curl http://localhost:8080/health

# 4) Open dashboard
# http://localhost:8080/dashboard
```

You can also run the core server directly:

```bash
make run-core
```

## OSS Commands (No Cloud Required)

- `clawhost` - Start core API server (default port `8080`)
- `clawhost help` - Show command usage
- `clawhost init` - Interactive setup wizard
- `clawhost deploy` - Deploy OpenClaw locally (self-managed mode)
- `clawhost upgrade` - Upgrade deployment metadata to latest OpenClaw version
- `clawhost destroy` - Remove everything (with confirmation)
- `clawhost status --all` - Show saved local deployment metadata
- `clawhost backup --output ./backup.json` - Export CLI state
- `clawhost restore --input ./backup.json` - Import CLI state

## Requested Command Flow

```bash
clawhost init                    # Interactive setup wizard
clawhost deploy                  # Deploy OpenClaw locally
clawhost upgrade                 # Upgrade to latest version
clawhost destroy                 # Remove everything (with confirmation)
```

## Optional Environment Variables

- `PORT` - API/dashboard port when running `clawhost` (default: `8080`)

Example:

```bash
PORT=9090 ./clawhost
```

## Local State Files

- `~/.clawhost/deployments.json`

## Notes

- `clawhost deploy` in OSS mode sets up local deployment state and checks local core health.
- `clawhost destroy` removes local deployment metadata and deletes `.clawhost.yaml` plus `~/.clawhost` after confirmation.
- `clawhost logs` still needs an API instance ID and is optional for basic local mode.

## Self-Managed Hosting (Your Own Server)

If you want to host this yourself (still open source):

1. Build on your server with `make build`
2. Run with `PORT=8080 ./clawhost`
3. Put Nginx/Caddy in front of it for HTTPS
4. Access the basic dashboard at `/dashboard`
