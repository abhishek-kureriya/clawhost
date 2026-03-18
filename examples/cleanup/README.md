# Cleanup Example

This tool helps you safely delete ClawHost instances when you're done with them.

## Usage

### Delete by Server ID

```bash
export HETZNER_API_TOKEN="your_token_here"
go run main.go -server-id="12345678"
```

### Delete by Server Name

```bash
export HETZNER_API_TOKEN="your_token_here"  
go run main.go -server-name="openclaw-demo-1234567890"
```

### Dry Run (Safe Testing)

```bash
# See what would be deleted without actually deleting
go run main.go -server-id="12345678" -dry-run
```

### Force Delete (No Confirmation)

```bash
# Skip confirmation prompt (useful for automation)
go run main.go -server-id="12345678" -force
```

## What This Tool Does

1. 🔍 **Looks up the server** by ID or name
2. 📋 **Shows server information** (IP, status, labels, etc.)
3. ⚠️ **Asks for confirmation** (unless -force is used)
4. ⏹️ **Stops the server** gracefully (if running)
5. 🗑️ **Deletes the server** permanently
6. 📝 **Provides cleanup suggestions** for related resources

## Safety Features

- **Dry run mode** to preview actions
- **Confirmation prompt** to prevent accidents  
- **Graceful shutdown** before deletion
- **Clear error messages** if something goes wrong
- **Detailed logging** of all actions

## Example Output

```
ClawHost Cleanup Tool
====================
🔍 Looking up server by ID: 12345678

📋 Server Information:
   ID: 12345678
   Name: openclaw-demo-1234567890
   Type: cx11
   Status: running
   Public IP: 78.46.123.456
   Location: nbg1
   Labels:
     purpose: openclaw-demo
     example: basic-deployment

⚠️  WARNING: This will permanently delete the server!
💾 All data on the server will be lost.
💰 You will stop being charged for this server.

Are you sure you want to delete 'openclaw-demo-1234567890' (12345678)? [y/N]: y
⏹️  Stopping server...
✅ Server stopped successfully
🗑️  Deleting server...

✅ Server deleted successfully!
💰 You will no longer be charged for this server.
📊 Check your Hetzner Cloud console to confirm deletion.

📝 Cleanup Suggestions:
   • Remove any DNS records pointing to 78.46.123.456
   • Delete any local SSH key entries for this server
   • Update any monitoring or backup systems
   • Remove firewall rules specific to this server

🎉 Cleanup complete!
```

## Troubleshooting

### "Server not found"
- Double-check the server ID or name
- Ensure you have the correct Hetzner API token
- The server may have already been deleted

### "Permission denied"
- Verify your Hetzner API token has delete permissions
- Check if the server is protected from deletion

### "Failed to stop server"
- The tool will still attempt deletion
- Check Hetzner console if the server gets stuck

## Cost Savings

💰 **Important**: Deleting servers immediately stops billing for:
- Server compute costs (~€4-20/month)
- Additional storage costs
- Network transfer overages

The server will be **immediately** removed from your bill.

## Related Commands

```bash
# List all your servers
hcloud server list

# Get server details  
hcloud server describe <server-id>

# Manual deletion via CLI
hcloud server delete <server-id>
```

## Automation

This tool is perfect for automation:

```bash
#!/bin/bash
# Cleanup script for temporary deployments

SERVERS=$(hcloud server list -o noheader -o columns=name | grep "temp-")

for server in $SERVERS; do
    echo "Deleting temporary server: $server"
    go run cleanup/main.go -server-name="$server" -force
done
```