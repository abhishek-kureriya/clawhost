package mcpbridge

import (
"bufio"
"context"
"encoding/json"
"fmt"
"io"
"log/slog"
"os"
"os/exec"
"sync"
"time"
)

// MCPServer represents a managed MCP subprocess.
type MCPServer struct {
Name      string
Command   string
Args      []string
EnvVars   map[string]string
logger    *slog.Logger
cmd       *exec.Cmd
stdin     io.WriteCloser
stdout    *bufio.Reader
stderr    *bufio.Reader
done      chan struct{}
healthy   bool
mu        sync.RWMutex
lastPing  time.Time
}

// NewMCPServer creates a new MCP server instance.
func NewMCPServer(name, command string, args []string, envVars map[string]string, logger *slog.Logger) *MCPServer {
if logger == nil {
logger = slog.Default()
}
return &MCPServer{
Name:     name,
Command:  command,
Args:     args,
EnvVars:  envVars,
logger:   logger,
done:     make(chan struct{}),
healthy:  false,
}
}

// Start begins the MCP subprocess and sets up communication channels.
func (s *MCPServer) Start(ctx context.Context) error {
s.mu.Lock()
defer s.mu.Unlock()

// Set up environment: inherit parent env + inject custom vars
s.cmd = exec.CommandContext(ctx, s.Command, s.Args...)
s.cmd.Env = os.Environ()
for k, v := range s.EnvVars {
s.cmd.Env = append(s.cmd.Env, fmt.Sprintf("%s=%s", k, v))
}

// Set up stdio
var err error
if s.stdin, err = s.cmd.StdinPipe(); err != nil {
return fmt.Errorf("stdin pipe error: %w", err)
}

stdout, err := s.cmd.StdoutPipe()
if err != nil {
return fmt.Errorf("stdout pipe error: %w", err)
}
s.stdout = bufio.NewReader(stdout)

stderr, err := s.cmd.StderrPipe()
if err != nil {
return fmt.Errorf("stderr pipe error: %w", err)
}
s.stderr = bufio.NewReader(stderr)

// Start the process
if err := s.cmd.Start(); err != nil {
return fmt.Errorf("failed to start MCP server %q: %w", s.Name, err)
}

s.healthy = true
s.lastPing = time.Now()
s.logger.Info("mcp_server_started", slog.String("name", s.Name))

// Monitor process in background
go s.monitorProcess()
go s.captureStderr()

return nil
}

// monitorProcess waits for the subprocess to exit and updates health status.
func (s *MCPServer) monitorProcess() {
err := s.cmd.Wait()

s.mu.Lock()
s.healthy = false
close(s.done)
s.mu.Unlock()

if err != nil {
s.logger.Error("mcp_server_exited", slog.String("name", s.Name), slog.Any("error", err))
} else {
s.logger.Info("mcp_server_exited", slog.String("name", s.Name))
}
}

// captureStderr logs stderr output from the MCP server.
func (s *MCPServer) captureStderr() {
for {
line, err := s.stderr.ReadString('\n')
if err != nil {
if err != io.EOF {
s.logger.Error("stderr read error", slog.String("name", s.Name), slog.Any("error", err))
}
return
}
s.logger.Debug("mcp_server_output", slog.String("name", s.Name), slog.String("message", line))
}
}

// SendRequest sends a JSON-RPC 2.0 request to the MCP server and reads the response.
func (s *MCPServer) SendRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
s.mu.RLock()
if !s.healthy || s.stdin == nil {
s.mu.RUnlock()
return nil, fmt.Errorf("MCP server %q is not healthy", s.Name)
}
s.mu.RUnlock()

// Marshal and send request
data, err := json.Marshal(req)
if err != nil {
return nil, fmt.Errorf("marshal error: %w", err)
}

// Add newline delimiter for line-based protocol
if _, err := s.stdin.Write(append(data, '\n')); err != nil {
s.mu.Lock()
s.healthy = false
s.mu.Unlock()
return nil, fmt.Errorf("write error: %w", err)
}

// Read response with timeout
respChan := make(chan *JSONRPCResponse, 1)
errChan := make(chan error, 1)

go func() {
line, err := s.stdout.ReadString('\n')
if err != nil {
errChan <- err
return
}

var resp JSONRPCResponse
if err := json.Unmarshal([]byte(line), &resp); err != nil {
errChan <- err
return
}

respChan <- &resp
}()

select {
case resp := <-respChan:
s.mu.Lock()
s.lastPing = time.Now()
s.mu.Unlock()
return resp, nil
case err := <-errChan:
s.mu.Lock()
s.healthy = false
s.mu.Unlock()
return nil, fmt.Errorf("response read error: %w", err)
case <-ctx.Done():
return nil, fmt.Errorf("request timeout")
}
}

// IsHealthy returns the current health status of the MCP server.
func (s *MCPServer) IsHealthy() bool {
s.mu.RLock()
defer s.mu.RUnlock()
return s.healthy
}

// Stop gracefully shuts down the MCP server.
func (s *MCPServer) Stop(ctx context.Context) error {
s.mu.Lock()
if s.stdin != nil {
s.stdin.Close()
}
s.healthy = false
s.mu.Unlock()

// Try graceful termination first
if s.cmd != nil && s.cmd.Process != nil {
s.cmd.Process.Signal(os.Interrupt)
}

select {
case <-s.done:
return nil
case <-ctx.Done():
// Force kill if graceful shutdown times out
if s.cmd != nil && s.cmd.Process != nil {
s.cmd.Process.Kill()
}
return ctx.Err()
}
}

// JSON-RPC 2.0 Request/Response structures

type JSONRPCRequest struct {
JSONRPC string          `json:"jsonrpc"`
ID      interface{}     `json:"id"`
Method  string          `json:"method"`
Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
JSONRPC string          `json:"jsonrpc"`
ID      interface{}     `json:"id"`
Result  json.RawMessage `json:"result,omitempty"`
Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
Code    int    `json:"code"`
Message string `json:"message"`
Data    string `json:"data,omitempty"`
}
