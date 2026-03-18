package provisioning

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// CloudProvider defines the interface for cloud providers
type CloudProvider interface {
	CreateServer(ctx context.Context, config ServerConfig) (*ServerInfo, error)
	GetServer(ctx context.Context, serverID string) (*ServerInfo, error)
	DeleteServer(ctx context.Context, serverID string) error
	WaitForServer(ctx context.Context, serverID string, targetStatus string, timeout time.Duration) error
}

type ServerConfig struct {
	Name       string
	ServerType string
	Location   string
	SSHKeyName string
	UserData   string
	Labels     map[string]string
}

type ServerInfo struct {
	ID        string
	Name      string
	Status    string
	PublicIP  string
	PrivateIP string
	Location  string
	Type      string
}

type HetznerProvider struct {
	client *hcloud.Client
}

func NewHetznerProvider(apiToken string) *HetznerProvider {
	client := hcloud.NewClient(hcloud.WithToken(apiToken))
	return &HetznerProvider{
		client: client,
	}
}

func (h *HetznerProvider) CreateServer(ctx context.Context, config ServerConfig) (*ServerInfo, error) {
	// Get server type
	serverType, _, err := h.client.ServerType.Get(ctx, config.ServerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get server type: %w", err)
	}

	// Get location
	location, _, err := h.client.Location.Get(ctx, config.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	// Get or create SSH key
	var sshKey *hcloud.SSHKey
	if config.SSHKeyName != "" {
		sshKey, _, err = h.client.SSHKey.Get(ctx, config.SSHKeyName)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH key: %w", err)
		}
	}

	// Get Ubuntu 22.04 image
	image, _, err := h.client.Image.GetForArchitecture(ctx, "ubuntu-22.04", serverType.Architecture)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ubuntu image: %w", err)
	}

	// Create server options
	opts := hcloud.ServerCreateOpts{
		Name:       config.Name,
		ServerType: serverType,
		Image:      image,
		Location:   location,
		UserData:   config.UserData,
		Labels:     config.Labels,
	}

	if sshKey != nil {
		opts.SSHKeys = []*hcloud.SSHKey{sshKey}
	}

	// Create server
	result, _, err := h.client.Server.Create(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	server := result.Server
	log.Printf("Created Hetzner server: %s (ID: %d)", server.Name, server.ID)

	return &ServerInfo{
		ID:       fmt.Sprintf("%d", server.ID),
		Name:     server.Name,
		Status:   string(server.Status),
		PublicIP: server.PublicNet.IPv4.IP.String(),
		Location: server.Datacenter.Location.Name,
		Type:     server.ServerType.Name,
	}, nil
}

func (h *HetznerProvider) GetServer(ctx context.Context, serverID string) (*ServerInfo, error) {
	server, _, err := h.client.Server.Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	if server == nil {
		return nil, fmt.Errorf("server not found")
	}

	var publicIP string
	if server.PublicNet.IPv4.IP != nil {
		publicIP = server.PublicNet.IPv4.IP.String()
	}

	var privateIP string
	if len(server.PrivateNet) > 0 && server.PrivateNet[0].IP != nil {
		privateIP = server.PrivateNet[0].IP.String()
	}

	return &ServerInfo{
		ID:        fmt.Sprintf("%d", server.ID),
		Name:      server.Name,
		Status:    string(server.Status),
		PublicIP:  publicIP,
		PrivateIP: privateIP,
		Location:  server.Datacenter.Location.Name,
		Type:      server.ServerType.Name,
	}, nil
}

func (h *HetznerProvider) DeleteServer(ctx context.Context, serverID string) error {
	server, _, err := h.client.Server.Get(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	if server == nil {
		return fmt.Errorf("server not found")
	}

	_, _, err = h.client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	log.Printf("Deleted Hetzner server: %s (ID: %s)", server.Name, serverID)
	return nil
}

func (h *HetznerProvider) WaitForServer(ctx context.Context, serverID string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		server, _, err := h.client.Server.Get(ctx, serverID)
		if err != nil {
			return fmt.Errorf("failed to get server status: %w", err)
		}

		if server == nil {
			return fmt.Errorf("server not found")
		}

		if string(server.Status) == targetStatus {
			return nil
		}

		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for server to reach status %s", targetStatus)
}
