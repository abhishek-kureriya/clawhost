package monitoring

import (
	"encoding/json"
	"time"
)

// Metric represents a single metric measurement
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// InstanceMetrics contains all metrics for an instance
type InstanceMetrics struct {
	InstanceID  string    `json:"instance_id"`
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIn   float64   `json:"network_in"`
	NetworkOut  float64   `json:"network_out"`
	Uptime      int64     `json:"uptime_seconds"`
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	CollectMetrics(instanceID string) (*InstanceMetrics, error)
	GetHistoricalMetrics(instanceID string, from, to time.Time) ([]*InstanceMetrics, error)
}

// BasicMetricsCollector implements basic metrics collection
type BasicMetricsCollector struct {
	metricsStore map[string][]*InstanceMetrics
}

func NewBasicMetricsCollector() *BasicMetricsCollector {
	return &BasicMetricsCollector{
		metricsStore: make(map[string][]*InstanceMetrics),
	}
}

func (c *BasicMetricsCollector) CollectMetrics(instanceID string) (*InstanceMetrics, error) {
	// In reality, this would collect actual metrics from the server
	// For now, we'll return mock data
	metrics := &InstanceMetrics{
		InstanceID:  instanceID,
		Timestamp:   time.Now(),
		CPUUsage:    45.2, // Mock CPU usage %
		MemoryUsage: 67.8, // Mock Memory usage %
		DiskUsage:   23.1, // Mock Disk usage %
		NetworkIn:   1024, // Mock Network in KB/s
		NetworkOut:  512,  // Mock Network out KB/s
		Uptime:      3600, // Mock uptime in seconds
	}

	// Store metrics
	if c.metricsStore[instanceID] == nil {
		c.metricsStore[instanceID] = make([]*InstanceMetrics, 0)
	}
	c.metricsStore[instanceID] = append(c.metricsStore[instanceID], metrics)

	// Keep only last 1000 metrics per instance
	if len(c.metricsStore[instanceID]) > 1000 {
		c.metricsStore[instanceID] = c.metricsStore[instanceID][1:]
	}

	return metrics, nil
}

func (c *BasicMetricsCollector) GetHistoricalMetrics(instanceID string, from, to time.Time) ([]*InstanceMetrics, error) {
	metrics := c.metricsStore[instanceID]
	if metrics == nil {
		return []*InstanceMetrics{}, nil
	}

	var filteredMetrics []*InstanceMetrics
	for _, metric := range metrics {
		if metric.Timestamp.After(from) && metric.Timestamp.Before(to) {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}

	return filteredMetrics, nil
}

// ToJSON converts metrics to JSON format
func (m *InstanceMetrics) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// HealthChecker interface for health checks
type HealthChecker interface {
	CheckHealth(instanceID string) (*HealthStatus, error)
}

// HealthStatus represents the health status of an instance
type HealthStatus struct {
	InstanceID string          `json:"instance_id"`
	Healthy    bool            `json:"healthy"`
	LastCheck  time.Time       `json:"last_check"`
	Message    string          `json:"message,omitempty"`
	Checks     map[string]bool `json:"checks"`
}

// BasicHealthChecker implements basic health checking
type BasicHealthChecker struct{}

func NewBasicHealthChecker() *BasicHealthChecker {
	return &BasicHealthChecker{}
}

func (h *BasicHealthChecker) CheckHealth(instanceID string) (*HealthStatus, error) {
	// In reality, this would perform actual health checks
	// For now, we'll return mock data
	checks := map[string]bool{
		"http_endpoint":    true,
		"database":         true,
		"disk_space":       true,
		"memory_available": true,
		"openclaw_process": true,
	}

	allHealthy := true
	for _, check := range checks {
		if !check {
			allHealthy = false
			break
		}
	}

	message := "All systems operational"
	if !allHealthy {
		message = "Some health checks failed"
	}

	return &HealthStatus{
		InstanceID: instanceID,
		Healthy:    allHealthy,
		LastCheck:  time.Now(),
		Message:    message,
		Checks:     checks,
	}, nil
}
