# ClawHost API Reference

Complete reference for the ClawHost REST API, covering both the open source core API and commercial service endpoints.

## Base URLs

- **Core API (Open Source)**: `http://localhost:8080`
- **Commercial Service**: `https://api.example.com`
- **Development**: `http://localhost:3000`

## Authentication

### Core API (Open Source)

The core API is **unauthenticated by default** for simplicity in self-hosted environments. For production use, implement your own authentication layer.

### Commercial Service API

Uses JWT authentication with API keys:

```bash
# Get your API key from the dashboard
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.example.com/api/v1/instances
```

**Authentication Headers:**
- `Authorization: Bearer <api_key>`
- `Content-Type: application/json`

## Core API Endpoints

### Health & Status

#### GET /health

Check the health status of the ClawHost API.

**Response:**
```json
{
  "status": "healthy",
  "service": "clawhost-core",
  "version": "1.0.0",
  "timestamp": "2026-03-18T10:30:00Z"
}
```

**Status Codes:**
- `200` - Service is healthy
- `503` - Service is unhealthy

#### GET /api/v1/version

Get detailed version information.

**Response:**
```json
{
  "version": "1.0.0",
  "commit": "abc123def456",
  "build_date": "2026-03-18T09:00:00Z",
  "go_version": "go1.23.0",
  "platform": "linux/amd64"
}
```

### Instance Management

#### GET /api/v1/instances

List all instances.

**Query Parameters:**
- `limit` (int): Maximum number of results (default: 50, max: 200)
- `offset` (int): Offset for pagination (default: 0)
- `status` (string): Filter by status (provisioning|running|stopped|error)
- `provider` (string): Filter by cloud provider (hetzner|aws|gcp)

**Response:**
```json
{
  "instances": [
    {
      "id": "inst-123abc",
      "name": "openclaw-demo",
      "status": "running",
      "provider": "hetzner",
      "server_type": "cx11",
      "location": "nbg1",
      "public_ip": "78.46.123.456",
      "private_ip": "10.0.1.5",
      "created_at": "2026-03-18T10:00:00Z",
      "updated_at": "2026-03-18T10:15:00Z",
      "labels": {
        "environment": "demo",
        "purpose": "testing"
      }
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

#### POST /api/v1/instances

Create a new OpenClaw instance.

**Request Body:**
```json
{
  "name": "my-openclaw-bot",
  "provider": "hetzner",
  "server_type": "cx11",
  "location": "nbg1",
  "openclaw_config": {
    "llm_provider": "openai",
    "llm_model": "gpt-3.5-turbo",
    "personality_prompt": "You are a helpful customer service assistant.",
    "business_knowledge": "We provide AI hosting services.",
    "api_keys": {
      "openai_api_key": "sk-..."
    }
  },
  "labels": {
    "environment": "production",
    "team": "customer-support"
  }
}
```

**Response:**
```json
{
  "instance_id": "inst-456def",
  "provisioning_job_id": "job-789ghi",
  "status": "provisioning",
  "estimated_duration": "5 minutes"
}
```

#### GET /api/v1/instances/{id}

Get detailed information about a specific instance.

**Path Parameters:**
- `id` (string): Instance ID

**Response:**
```json
{
  "id": "inst-123abc",
  "name": "openclaw-demo",
  "status": "running",
  "provider": "hetzner",
  "server_id": "htz-789xyz",
  "server_type": "cx11",
  "location": "nbg1",
  "public_ip": "78.46.123.456",
  "private_ip": "10.0.1.5",
  "openclaw_config": {
    "llm_provider": "openai",
    "llm_model": "gpt-3.5-turbo",
    "personality_prompt": "You are a helpful assistant.",
    "version": "1.0.0",
    "status": "healthy"
  },
  "resource_usage": {
    "cpu_cores": 1,
    "memory_gb": 4,
    "disk_gb": 20,
    "bandwidth_gb": 1000
  },
  "costs": {
    "monthly_estimate": 4.15,
    "currency": "EUR"
  },
  "created_at": "2026-03-18T10:00:00Z",
  "updated_at": "2026-03-18T10:15:00Z",
  "labels": {
    "environment": "demo"
  }
}
```

#### PUT /api/v1/instances/{id}

Update an instance configuration.

**Request Body:**
```json
{
  "name": "updated-openclaw-bot",
  "openclaw_config": {
    "personality_prompt": "Updated assistant personality.",
    "business_knowledge": "Updated business context."
  },
  "labels": {
    "environment": "production",
    "version": "2.0"
  }
}
```

#### DELETE /api/v1/instances/{id}

Delete an instance permanently.

**Query Parameters:**
- `force` (boolean): Force deletion even if instance is running
- `backup` (boolean): Create backup before deletion (commercial feature)

**Response:**
```json
{
  "message": "Instance deletion initiated",
  "job_id": "job-delete-123",
  "estimated_duration": "2 minutes"
}
```

### Instance Operations

#### POST /api/v1/instances/{id}/start

Start a stopped instance.

**Response:**
```json
{
  "message": "Instance start initiated",
  "job_id": "job-start-456"
}
```

#### POST /api/v1/instances/{id}/stop

Stop a running instance.

**Request Body (optional):**
```json
{
  "graceful": true,
  "timeout_seconds": 30
}
```

#### POST /api/v1/instances/{id}/restart

Restart an instance.

**Request Body (optional):**
```json
{
  "graceful": true,
  "timeout_seconds": 30
}
```

### Monitoring & Metrics

#### GET /api/v1/instances/{id}/status

Get current status of an instance.

**Response:**
```json
{
  "instance_id": "inst-123abc",
  "status": "running",
  "health": "healthy",
  "uptime": "2h 15m 30s",
  "last_check": "2026-03-18T10:30:00Z",
  "openclaw_version": "1.0.0",
  "response_time_ms": 45
}
```

**Status Values:**
- `provisioning` - Being created
- `starting` - Starting up
- `running` - Operational
- `stopping` - Shutting down  
- `stopped` - Not running
- `error` - Error state
- `maintenance` - Under maintenance

#### GET /api/v1/instances/{id}/metrics

Get real-time performance metrics.

**Query Parameters:**
- `period` (string): Time period (1h|6h|24h|7d|30d)
- `resolution` (string): Data resolution (1m|5m|1h)

**Response:**
```json
{
  "instance_id": "inst-123abc",
  "period": "1h",
  "resolution": "5m",
  "metrics": {
    "cpu": {
      "current": 45.2,
      "average": 38.7,
      "peak": 78.1,
      "unit": "percent"
    },
    "memory": {
      "current": 67.8,
      "available_gb": 1.3,
      "unit": "percent"
    },
    "disk": {
      "used_gb": 4.6,
      "available_gb": 15.4,
      "usage_percent": 23.1
    },
    "network": {
      "incoming_kbps": 1024,
      "outgoing_kbps": 512,
      "total_gb": 12.5
    },
    "openclaw": {
      "requests_per_minute": 45,
      "response_time_ms": 150,
      "error_rate": 0.1,
      "active_conversations": 12
    }
  },
  "timestamp": "2026-03-18T10:30:00Z"
}
```

#### GET /api/v1/instances/{id}/health

Perform comprehensive health check.

**Response:**
```json
{
  "instance_id": "inst-123abc",
  "healthy": true,
  "score": 95,
  "last_check": "2026-03-18T10:30:00Z",
  "checks": {
    "http_endpoint": {
      "status": "pass",
      "response_time_ms": 45,
      "details": "OpenClaw API responding normally"
    },
    "database": {
      "status": "pass",
      "connection_time_ms": 12,
      "details": "Database connection healthy"
    },
    "disk_space": {
      "status": "pass",
      "free_percent": 77,
      "details": "15.4GB available"
    },
    "memory": {
      "status": "warn",
      "usage_percent": 68,
      "details": "Memory usage approaching threshold"
    },
    "ssl_certificate": {
      "status": "pass",
      "expires_in_days": 75,
      "details": "Certificate valid until 2026-06-01"
    }
  }
}
```

**Health Check Status:**
- `pass` - Check passed
- `warn` - Warning condition
- `fail` - Check failed

#### GET /api/v1/instances/{id}/logs

Retrieve instance logs.

**Query Parameters:**
- `limit` (int): Number of entries (default: 100, max: 1000)
- `level` (string): Filter by level (debug|info|warn|error)
- `since` (string): ISO 8601 timestamp
- `until` (string): ISO 8601 timestamp
- `search` (string): Search term in log messages
- `source` (string): Filter by source (openclaw|system|nginx)

**Response:**
```json
{
  "instance_id": "inst-123abc",
  "total_entries": 1247,
  "returned_entries": 100,
  "logs": [
    {
      "timestamp": "2026-03-18T10:30:15Z",
      "level": "info",
      "source": "openclaw",
      "message": "Processed user request successfully",
      "request_id": "req-789xyz",
      "user_id": "user-456abc",
      "response_time_ms": 145
    },
    {
      "timestamp": "2026-03-18T10:30:10Z",
      "level": "debug",
      "source": "system",
      "message": "Memory usage: 68%",
      "cpu_usage": 45.2,
      "memory_mb": 2765
    }
  ]
}
```

### Provisioning

#### GET /api/v1/provisioning/jobs

List provisioning jobs.

**Query Parameters:**
- `limit` (int): Maximum results
- `offset` (int): Pagination offset
- `status` (string): Filter by status
- `instance_id` (string): Filter by instance

**Response:**
```json
{
  "jobs": [
    {
      "id": "job-123abc",
      "instance_id": "inst-456def",
      "type": "create_instance",
      "status": "completed",
      "progress": 100,
      "current_step": "Instance ready",
      "started_at": "2026-03-18T10:00:00Z",
      "completed_at": "2026-03-18T10:15:00Z",
      "duration_seconds": 900
    }
  ]
}
```

#### GET /api/v1/provisioning/jobs/{job_id}

Get detailed job status.

**Response:**
```json
{
  "id": "job-123abc",
  "instance_id": "inst-456def",
  "type": "create_instance",
  "status": "in_progress",
  "progress": 75,
  "current_step": "Installing OpenClaw",
  "started_at": "2026-03-18T10:00:00Z",
  "estimated_completion": "2026-03-18T10:15:00Z",
  "steps": [
    {
      "name": "Initializing",
      "status": "completed",
      "started_at": "2026-03-18T10:00:00Z",
      "completed_at": "2026-03-18T10:00:30Z"
    },
    {
      "name": "Creating server",
      "status": "completed",
      "started_at": "2026-03-18T10:00:30Z",
      "completed_at": "2026-03-18T10:05:00Z"
    },
    {
      "name": "Installing OpenClaw",
      "status": "in_progress",
      "started_at": "2026-03-18T10:05:00Z",
      "progress": 60
    },
    {
      "name": "Configuring services",
      "status": "pending"
    }
  ],
  "logs": [
    "2026-03-18T10:00:00Z: Job started",
    "2026-03-18T10:00:30Z: Server creation initiated",
    "2026-03-18T10:05:00Z: Server ready, installing OpenClaw",
    "2026-03-18T10:12:00Z: OpenClaw installation 60% complete"
  ]
}
```

### Cloud Providers

#### GET /api/v1/providers

List available cloud providers.

**Response:**
```json
{
  "providers": [
    {
      "name": "hetzner",
      "display_name": "Hetzner Cloud",
      "regions": [
        {
          "id": "nbg1",
          "name": "Nuremberg",
          "country": "Germany"
        },
        {
          "id": "fsn1",
          "name": "Falkenstein",
          "country": "Germany"
        }
      ],
      "server_types": [
        {
          "id": "cx11",
          "name": "CX11",
          "cores": 1,
          "memory": 4,
          "disk": 20,
          "price_monthly": 4.15,
          "currency": "EUR"
        }
      ]
    }
  ]
}
```

## Commercial Service API

*Additional endpoints available in the commercial service:*

### Customer Management

#### GET /api/v1/customers/me
Get current customer information.

#### PUT /api/v1/customers/me
Update customer profile.

### Billing

#### GET /api/v1/billing/subscription
Get subscription details.

#### POST /api/v1/billing/subscription
Update subscription plan.

#### GET /api/v1/billing/invoices
List invoices.

### Support

#### GET /api/v1/support/tickets
List support tickets.

#### POST /api/v1/support/tickets
Create support ticket.

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "INSTANCE_NOT_FOUND",
    "message": "Instance with ID 'inst-123' not found",
    "details": {
      "instance_id": "inst-123",
      "suggestion": "Check the instance ID and try again"
    },
    "request_id": "req-456def789",
    "timestamp": "2026-03-18T10:30:00Z"
  }
}
```

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `202` - Accepted (async operation)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `429` - Rate Limited
- `500` - Internal Server Error
- `503` - Service Unavailable

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Request validation failed |
| `AUTHENTICATION_REQUIRED` | Missing or invalid authentication |
| `INSUFFICIENT_PERMISSIONS` | User lacks required permissions |
| `INSTANCE_NOT_FOUND` | Instance does not exist |
| `INSTANCE_NOT_READY` | Instance not ready for operation |
| `PROVISIONING_FAILED` | Server provisioning failed |
| `QUOTA_EXCEEDED` | Account quota exceeded |
| `PAYMENT_REQUIRED` | Payment issue (commercial) |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Unexpected server error |

## Rate Limiting

### Limits

- **Core API (Open Source)**: 100 requests/minute per IP
- **Commercial API**: 1000 requests/minute per API key
- **Provisioning Operations**: 10 requests/minute

### Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 987
X-RateLimit-Reset: 1642781234
X-RateLimit-Window: 60
```

### Rate Limit Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "details": {
      "limit": 1000,
      "window_seconds": 60,
      "reset_at": "2026-03-18T10:31:00Z"
    }
  }
}
```

## Webhooks

### Event Types

- `instance.created` - New instance created
- `instance.started` - Instance started
- `instance.stopped` - Instance stopped
- `instance.deleted` - Instance deleted
- `instance.health.degraded` - Health check failed
- `provisioning.completed` - Provisioning finished
- `provisioning.failed` - Provisioning failed
- `billing.payment.succeeded` - Payment successful (commercial)
- `billing.payment.failed` - Payment failed (commercial)

### Webhook Payload

```json
{
  "id": "evt-123abc",
  "type": "instance.created",
  "created": "2026-03-18T10:30:00Z",
  "data": {
    "instance": {
      "id": "inst-456def",
      "name": "openclaw-demo",
      "status": "provisioning",
      "provider": "hetzner"
    }
  }
}
```

## SDKs and Examples

### Go SDK

```go
import "github.com/yourusername/clawhost-go"

client := clawhost.NewClient("http://localhost:8080")
instances, err := client.ListInstances(ctx, &clawhost.ListOptions{
    Limit: 10,
    Status: "running",
})
```

### JavaScript/Node.js

```javascript
const ClawHost = require('clawhost-js');

const client = new ClawHost({
    baseURL: 'http://localhost:8080',
    apiKey: 'your-api-key'
});

const instances = await client.instances.list({
    limit: 10,
    status: 'running'
});
```

### Python

```python
import clawhost

client = clawhost.Client(
    base_url='http://localhost:8080',
    api_key='your-api-key'
)

instances = client.instances.list(limit=10, status='running')
```

### cURL Examples

```bash
# Health check
curl -X GET http://localhost:8080/health

# List instances
curl -X GET "http://localhost:8080/api/v1/instances?limit=10"

# Create instance
curl -X POST http://localhost:8080/api/v1/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-bot",
    "provider": "hetzner",
    "server_type": "cx11",
    "openclaw_config": {
      "llm_provider": "openai",
      "personality_prompt": "You are helpful."
    }
  }'

# Get instance status
curl -X GET http://localhost:8080/api/v1/instances/inst-123/status

# Get metrics
curl -X GET "http://localhost:8080/api/v1/instances/inst-123/metrics?period=1h"
```

---

**📚 Complete API Reference**

*This reference covers all available endpoints. For additional examples and tutorials, see the [examples/](../examples/) directory and [community resources](../community/).*