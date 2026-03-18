# Open Source CLI (Local / Self-Managed)

This guide is for the **open source core only**.

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
- `clawhost init --name local` - Create a local deployment profile/state
- `clawhost status --all` - Show saved local deployment metadata
- `clawhost backup --output ./backup.json` - Export CLI state
- `clawhost restore --input ./backup.json` - Import CLI state

## Optional Environment Variables

- `PORT` - API/dashboard port when running `clawhost` (default: `8080`)

Example:

```bash
PORT=9090 ./clawhost
```

## Local State Files

- `~/.clawhost/deployments.json`

## Commands Intentionally Excluded From This OSS-Local Guide

The following commands are for cloud provisioning flows and are not needed for simple self-managed local mode:

- `clawhost deploy`
- `clawhost destroy`
- `clawhost logs` (depends on provisioned instance IDs)

## Self-Managed Hosting (Your Own Server)

If you want to host this yourself (still open source):

1. Build on your server with `make build`
2. Run with `PORT=8080 ./clawhost`
3. Put Nginx/Caddy in front of it for HTTPS
4. Access the basic dashboard at `/dashboard`
