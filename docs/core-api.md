# ClawHost Core API Reference

The ClawHost Core API provides the foundational endpoints for managing OpenClaw instances. This is the **open source API** that powers the core functionality.

## Base URL

```
http://localhost:8080  # Default local development
```

## Authentication

The core API is **unauthenticated by default** for simplicity. For production use, implement your own authentication layer or use the commercial hosting service.

## Health Check

### GET /health

Returns the health status of the ClawHost Core API.

#### Response

```json
{
  "status": "healthy",
  "service": "clawhost-core",
  "version": "1.0.0"
}
```

## Instance Management

### GET /api/v1/instances/{id}/status

Get the current status of an OpenClaw instance.

#### Parameters

- `id` (string, required): The instance ID

#### Response

```json
{
  "instance_id": "inst-123",
  "status": "running",
  "uptime": "2h 15m",
  "version": "openclaw-1.0.0"
}
```

#### Status Values

- `provisioning`: Instance is being created
- `starting`: Instance is starting up
- `running`: Instance is operational
- `stopping`: Instance is shutting down
- `stopped`: Instance is stopped
- `error`: Instance has encountered an error

### GET /api/v1/instances/{id}/metrics

Get real-time metrics for an instance.

#### Response

```json
{
  "instance_id": "inst-123",
  "cpu_usage": 45.2,
  "memory_usage": 67.8,
  "disk_usage": 23.1,
  "network_in": 1024,
  "network_out": 512,
  "timestamp": "2026-03-18T10:30:00Z"
}
```

#### Metrics Description

- `cpu_usage`: CPU utilization percentage (0-100)
- `memory_usage`: Memory utilization percentage (0-100)
- `disk_usage`: Disk utilization percentage (0-100)
- `network_in`: Incoming network traffic (KB/s)
- `network_out`: Outgoing network traffic (KB/s)

### GET /api/v1/instances/{id}/health

Perform a comprehensive health check on an instance.

#### Response

```json
{
  "instance_id": "inst-123",
  "healthy": true,
  "last_check": "2026-03-18T10:30:00Z",
  "checks": {
    "http_endpoint": true,
    "database": true,
    "disk_space": true,
    "memory_available": true
  }
}
```

#### Health Checks

- `http_endpoint`: OpenClaw HTTP API is responding
- `database`: Database connection is healthy
- `disk_space`: Sufficient disk space available (>10% free)
- `memory_available`: Sufficient memory available (>20% free)

### GET /api/v1/instances/{id}/logs

Retrieve recent logs from an instance.

#### Query Parameters

- `limit` (integer, optional): Number of log entries to return (default: 100, max: 1000)
- `level` (string, optional): Filter by log level (debug, info, warn, error)
- `since` (string, optional): ISO 8601 timestamp to get logs after

#### Example Request

```bash
GET /api/v1/instances/inst-123/logs?limit=50&level=error&since=2026-03-18T09:00:00Z
```

#### Response

```json
{
  "instance_id": "inst-123",
  "limit": 100,
  "logs": [
    {
      "timestamp": "2026-03-18T10:30:00Z",
      "level": "info",
      "message": "OpenClaw instance started",
      "source": "openclaw"
    },
    {
      "timestamp": "2026-03-18T10:29:45Z",
      "level": "info",
      "message": "Database connection established",
      "source": "system"
    }
  ]
}
```

## Provisioning

### GET /api/v1/provision/status/{job_id}

Get the status of a provisioning job.

#### Parameters

- `job_id` (string, required): The provisioning job ID

#### Response

```json
{
  "job_id": "job-456",
  "status": "completed",
  "progress": 100,
  "current_step": "Instance ready",
  "started_at": "2026-03-18T10:00:00Z",
  "completed_at": "2026-03-18T10:15:00Z"
}
```

#### Status Values

- `pending`: Job is queued
- `in_progress`: Job is executing
- `completed`: Job finished successfully
- `failed`: Job encountered an error
- `cancelled`: Job was cancelled

#### Progress Steps

1. `Initializing`: Job setup
2. `Creating server`: Provisioning cloud infrastructure
3. `Installing OpenClaw`: Installing and configuring software
4. `Configuring services`: Setting up monitoring, backups, etc.
5. `Instance ready`: Deployment complete

## Error Handling

### Error Response Format

```json
{
  "error": "Instance not found",
  "code": "INSTANCE_NOT_FOUND",
  "details": {
    "instance_id": "inst-123"
  }
}
```

### HTTP Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Error Codes

- `INSTANCE_NOT_FOUND`: Instance ID does not exist
- `INSTANCE_NOT_READY`: Instance is not ready for the requested operation
- `PROVISIONING_FAILED`: Server provisioning failed
- `METRICS_UNAVAILABLE`: Unable to collect metrics
- `INVALID_PARAMETERS`: Request parameters are invalid

## Rate Limiting

The core API has basic rate limiting:

- **100 requests per minute** per IP address
- **1000 requests per hour** per IP address

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642781234
```

## Webhooks

### Instance Events

The core API can send webhooks for instance events. Configure webhook URLs in your environment:

```bash
export WEBHOOK_URL="https://your-app.com/webhooks/clawhost"
```

#### Event Types

- `instance.status.changed`: Instance status changed
- `instance.health.degraded`: Health check failed
- `instance.metrics.alert`: Metrics threshold exceeded
- `provisioning.completed`: Provisioning finished
- `provisioning.failed`: Provisioning failed

#### Webhook Payload

```json
{
  "event": "instance.status.changed",
  "timestamp": "2026-03-18T10:30:00Z",
  "data": {
    "instance_id": "inst-123",
    "old_status": "starting",
    "new_status": "running"
  }
}
```

## SDK and Client Libraries

### Official Go SDK

```bash
go get github.com/yourusername/clawhost-go-sdk
```

```go
import "github.com/yourusername/clawhost-go-sdk"

client := clawhost.NewClient("http://localhost:8080")
status, err := client.GetInstanceStatus("inst-123")
```

### Community SDKs

- **Node.js**: `npm install clawhost-node-sdk`
- **Python**: `pip install clawhost-python-sdk`
- **PHP**: `composer require clawhost/php-sdk`

## Examples

### Monitoring Script

```bash
#!/bin/bash
# Simple monitoring script

INSTANCE_ID="inst-123"
API_BASE="http://localhost:8080/api/v1"

# Check health
health=$(curl -s "$API_BASE/instances/$INSTANCE_ID/health")
if [[ $(echo $health | jq -r '.healthy') == "true" ]]; then
    echo "✅ Instance is healthy"
else
    echo "❌ Instance health check failed"
    echo $health | jq '.checks'
fi

# Get metrics
metrics=$(curl -s "$API_BASE/instances/$INSTANCE_ID/metrics")
cpu=$(echo $metrics | jq -r '.cpu_usage')
memory=$(echo $metrics | jq -r '.memory_usage')

echo "CPU Usage: ${cpu}%"
echo "Memory Usage: ${memory}%"
```

### Provisioning Monitor

```python
import requests
import time
import sys

def monitor_provisioning(job_id):
    base_url = "http://localhost:8080/api/v1"
    
    while True:
        response = requests.get(f"{base_url}/provision/status/{job_id}")
        data = response.json()
        
        print(f"Status: {data['status']} - {data['current_step']} ({data['progress']}%)")
        
        if data['status'] in ['completed', 'failed', 'cancelled']:
            break
            
        time.sleep(10)
    
    return data['status'] == 'completed'

if __name__ == "__main__":
    job_id = sys.argv[1]
    success = monitor_provisioning(job_id)
    print(f"Provisioning {'succeeded' if success else 'failed'}")
```

## Next Steps

- **[Provisioning Guide](provisioning.md)** - Learn how to deploy instances
- **[Monitoring Setup](monitoring.md)** - Set up advanced monitoring
- **[Examples](../examples/)** - Complete usage examples
- **[Community](../community/)** - Get help and share ideas

---

*The ClawHost Core API is open source and free to use. For managed hosting with additional features, check out [ClawHost Commercial Service](https://clawhost.com).*