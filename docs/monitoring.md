# Monitoring Guide

This guide covers how to monitor ClawHost instances and detect issues early.

## What to Monitor

- Instance health status
- CPU, memory, disk usage
- Network throughput
- OpenClaw request latency and error rate
- Provisioning success/failure rate

## Core Endpoints

- `GET /health` - API process health
- `GET /api/v1/instances/{id}/status` - Instance lifecycle status
- `GET /api/v1/instances/{id}/health` - Health checks
- `GET /api/v1/instances/{id}/metrics` - Runtime metrics
- `GET /api/v1/instances/{id}/logs` - Logs for debugging

Endpoint payload details are documented in:

- [Core API](core-api.md)
- [Full API Reference](api-reference.md)

## Alerting Baseline

Recommended first alerts:

- Instance not healthy for more than 5 minutes
- CPU over 85% for 10+ minutes
- Memory over 90% for 5+ minutes
- Disk free below 15%
- Error rate above 2%

## Log Monitoring

Track and retain logs for:

- Provisioning jobs
- OpenClaw service errors
- Authentication and access events
- API errors and timeouts

## Production Checklist

- Health checks run continuously.
- Alerts route to on-call channel.
- Dashboard includes all active instances.
- Log retention policy is defined.
- Weekly review of incidents and trends.
