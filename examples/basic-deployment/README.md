# Basic Deployment Example

This example shows how to deploy a single OpenClaw instance using the ClawHost core.

## Prerequisites

1. **Hetzner Cloud Account**: [Sign up here](https://hetzner.com/cloud)
2. **API Token**: Generate in Hetzner Cloud Console
3. **Go 1.23+**: [Download here](https://golang.org/dl/)

## Quick Start

1. **Set your API token**:
   ```bash
   export HETZNER_API_TOKEN="your_hetzner_token_here"
   ```

2. **Run the example**:
   ```bash
   go run main.go
   ```

3. **Wait for deployment** (3-5 minutes total):
   - Server creation: ~30 seconds
   - OpenClaw installation: 2-4 minutes

4. **Access your bot**:
   ```bash
   # The script will output the IP address
   curl http://YOUR_SERVER_IP/health
   ```

## What This Example Does

1. 🎆 **Creates a Hetzner Cloud server** (cx11 - 1 vCPU, 4GB RAM)
2. 🤖 **Installs OpenClaw** automatically via cloud-init
3. 🎨 **Configures the AI personality** for customer support
4. 🌐 **Sets up Nginx reverse proxy** for web access
5. ✅ **Provides health check endpoint**

## Configuration Options

Edit the `openclawConfig` in `main.go`:

```go
openclawConfig := provisioning.OpenClawConfig{
    LLMProvider:       "openai",           // openai, anthropic, google
    LLMModel:          "gpt-3.5-turbo",    // gpt-4, claude-3-sonnet, etc.
    PersonalityPrompt: "Your custom prompt here",
    BusinessKnowledge: "Context about your business",
}
```

## Server Types

Choose different Hetzner server types:

| Type | vCPUs | RAM | Disk | Price/month |
|------|-------|-----|------|-------------|
| cx11 | 1 | 4GB | 20GB | ~€4 |
| cx21 | 2 | 8GB | 40GB | ~€8 |
| cx31 | 4 | 16GB | 80GB | ~€16 |

## Monitoring Your Instance

After deployment, monitor your instance:

```bash
# Check instance status
curl http://localhost:8080/api/v1/instances/YOUR_INSTANCE_ID/status

# Get metrics
curl http://localhost:8080/api/v1/instances/YOUR_INSTANCE_ID/metrics

# Health check
curl http://YOUR_SERVER_IP/health
```

## Cleanup

Don't forget to clean up when done:

```bash
# Using the cleanup example
go run ../cleanup/main.go -server-id=YOUR_SERVER_ID

# Or via Hetzner console
hcloud server delete YOUR_SERVER_NAME
```

## Troubleshooting

### Server Creation Fails
- Check your Hetzner API token
- Verify you have quota for the server type
- Try a different location (fsn1, hel1)

### OpenClaw Not Responding
- Wait longer (initial setup takes 3-5 minutes)
- SSH to server and check logs: `sudo journalctl -u docker`
- Verify cloud-init completed: `sudo cloud-init status`

### API Errors
- Ensure core API server is running: `go run ../../core/cmd/main.go`
- Check server logs for detailed error messages

## Next Steps

- 📚 **[Multi-Instance Example](../multi-instance/)** - Deploy multiple bots
- 🔧 **[Custom Provider Example](../custom-provider/)** - Add new cloud providers
- 📈 **[Monitoring Example](../monitoring/)** - Set up advanced monitoring

## Estimated Costs

- **Server**: €4/month (cx11)
- **Bandwidth**: Usually free up to 1TB
- **Total**: ~€4/month for a basic setup

**Much cheaper than managed alternatives at €49-199/month!** 💰