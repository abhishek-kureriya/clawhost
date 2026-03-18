package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yourusername/clawhost/core/provisioning"
)

func main() {
	fmt.Println("ClawHost Cleanup Tool")
	fmt.Println("====================")

	var (
		serverID   = flag.String("server-id", "", "Server ID to delete")
		serverName = flag.String("server-name", "", "Server name to delete")
		dryRun     = flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
		force      = flag.Bool("force", false, "Skip confirmation prompt")
	)
	flag.Parse()

	if *serverID == "" && *serverName == "" {
		fmt.Println("❌ Error: Either -server-id or -server-name must be provided")
		flag.PrintDefaults()
		os.Exit(1)
	}

	apiToken := os.Getenv("HETZNER_API_TOKEN")
	if apiToken == "" {
		log.Fatal("❌ Please set HETZNER_API_TOKEN environment variable")
	}

	provider := provisioning.NewHetznerProvider(apiToken)
	ctx := context.Background()

	var server *provisioning.ServerInfo
	var err error

	if *serverID != "" {
		fmt.Printf("🔍 Looking up server by ID: %s\n", *serverID)
		server, err = provider.GetServer(ctx, *serverID)
	} else {
		log.Fatalf("❌ Deleting by name is not yet supported in this example; use -server-id")
	}

	if err != nil {
		log.Fatalf("❌ Failed to get server: %v", err)
	}

	fmt.Printf("\n📋 Server Information:\n")
	fmt.Printf("   ID: %s\n", server.ID)
	fmt.Printf("   Name: %s\n", server.Name)
	fmt.Printf("   Type: %s\n", server.Type)
	fmt.Printf("   Status: %s\n", server.Status)
	fmt.Printf("   Public IP: %s\n", server.PublicIP)
	fmt.Printf("   Location: %s\n", server.Location)

	if *dryRun {
		fmt.Printf("\n🔍 DRY RUN MODE - No changes will be made\n")
		fmt.Printf("✅ Would delete server: %s (%s)\n", server.Name, server.ID)
		fmt.Printf("💰 This would stop billing for this server\n")
		if server.Status == "running" {
			fmt.Printf("⚠️  Server is currently running and would be powered off\n")
		}
		return
	}

	if !*force {
		fmt.Printf("\n⚠️  WARNING: This will permanently delete the server!\n")
		fmt.Printf("💾 All data on the server will be lost.\n")
		fmt.Printf("💰 You will stop being charged for this server.\n")
		fmt.Printf("\nAre you sure you want to delete '%s' (%s)? [y/N]: ", server.Name, server.ID)

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("❌ Deletion cancelled.")
			return
		}
	}

	fmt.Printf("🗑️  Deleting server...\n")
	if err := provider.DeleteServer(ctx, server.ID); err != nil {
		log.Fatalf("❌ Failed to delete server: %v", err)
	}

	fmt.Printf("\n✅ Server deleted successfully!\n")
	fmt.Printf("💰 You will no longer be charged for this server.\n")
	fmt.Printf("📊 Check your Hetzner Cloud console to confirm deletion.\n")
}
