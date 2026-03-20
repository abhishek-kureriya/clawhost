# Hosting Service (Paid Version)

This module is the commercial backend for ClawHost paid hosting.

## What is included

- Customer dashboard API endpoints (database-backed)
- Billing API integration (Stripe)
- Stripe webhook sync for subscription status updates
- Support ticket API endpoints (database-backed)
- JWT authentication for paid API routes
- Core provisioning bridge for paid instance creation

## Run locally

From repository root:

```bash
make build-hosting
make run-hosting
```

Default port: `8090` (override with `HOSTING_PORT`)

Database:

- `HOSTING_DATABASE_URL` for PostgreSQL
- or fallback SQLite via `HOSTING_SQLITE_PATH` (default `hosting-service.db`)

Auth:

- Set `HOSTING_JWT_SECRET` for JWT auth
- Optional local helper: `ALLOW_DEV_AUTH=true` enables `/api/v1/auth/dev-token`

Core integration:

- Set `CORE_API_URL` so paid instance creation can call core provisioning API

Webhook:

- Set `STRIPE_WEBHOOK_SECRET` for `/webhooks/stripe`

## Health check

```bash
curl http://localhost:8090/health
```

## Example authenticated request

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8090/api/v1/dashboard
```

## Get a local dev token (optional)

```bash
curl -X POST http://localhost:8090/api/v1/auth/dev-token \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","name":"Dev User"}'
```

## Billing note

Set `STRIPE_SECRET_KEY` to enable billing endpoints. If not set, billing routes return `503` with configuration guidance.
