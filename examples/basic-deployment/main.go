package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/clawhost/core/provisioning"
)

func main() {
	fmt.Println("ClawHost Basic Deployment Example")
	fmt.Println("==================================")

	// Get Hetzner API token from environment
	apiToken := os.Getenv("HETZNER_API_TOKEN")
	if apiToken == "" {
		log.Fatal("Please set HETZNER_API_TOKEN environment variable")
	}

	// Create Hetzner provider
	provider := provisioning.NewHetznerProvider(apiToken)

	// Configure OpenClaw instance
	openclawConfig := provisioning.OpenClawConfig{
		LLMProvider:       "openai",
		LLMModel:          "gpt-3.5-turbo",
		PersonalityPrompt: "You are a friendly AI assistant for customer support.",
		BusinessKnowledge: "We are a company that provides AI hosting services.",
	}

	// Generate cloud-init script
	userData := provisioning.GenerateCloudInitScript(openclawConfig)

	// Server configuration
	serverConfig := provisioning.ServerConfig{
		Name:       fmt.Sprintf("openclaw-demo-%d", time.Now().Unix()),
		ServerType: "cx11", // Small server for demo
		Location:   "nbg1", // Nuremberg data center
		UserData:   userData,
		Labels: map[string]string{
			"purpose": "openclaw-demo",
			"example": "basic-deployment",
		},
	}

	ctx := context.Background()

	fmt.Println("Creating server...")
	server, err := provider.CreateServer(ctx, serverConfig)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Server created successfully!\n")
	fmt.Printf("ID: %s\n", server.ID)
	fmt.Printf("Name: %s\n", server.Name)
	fmt.Printf("Public IP: %s\n", server.PublicIP)
	fmt.Printf("Status: %s\n", server.Status)

	fmt.Println("\nWaiting for server to start...")
	err = provider.WaitForServer(ctx, server.ID, "running", 5*time.Minute)
	if err != nil {
		log.Printf("Warning: Server may not be fully ready: %v", err)
	} else {
		fmt.Println("Server is now running!")
	}

	fmt.Printf("\n✅ Deployment Complete!\n")
	fmt.Printf("OpenClaw will be available at: http://%s\n", server.PublicIP)
	fmt.Printf("Please allow 3-5 minutes for OpenClaw to fully initialize.\n")
	fmt.Printf("\nTo monitor your instance:")
	fmt.Printf("curl http://localhost:8080/api/v1/instances/%s/status\n", server.ID)
	fmt.Printf("\nTo delete this server later:")
	fmt.Printf("go run ../cleanup/main.go -server-id=%s\n", server.ID)
}
