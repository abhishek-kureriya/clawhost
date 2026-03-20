# Hosting Service Architecture

## Product Flow

```
┌─ Marketing Site ─┐    ┌─ Onboarding App ─┐    ┌─ Customer Dashboard ─┐
│ • Landing page   │    │ • Sign up wizard │    │ • Manage OpenClaw    │
│ • Pricing        │ -> │ • Payment setup  │ -> │ • Analytics          │
│ • Testimonials   │    │ • AI configuration│   │ • Platform settings  │
└──────────────────┘    └──────────────────┘    └──────────────────────┘
        │
        ▼
      ┌─ Provisioning API ─┐
      │ • Create server     │
      │ • Install OpenClaw  │
      │ • Connect platforms │
      └─────────────────────┘
```

## Repository Structure

### Open Source Core (`/core/`)
```
core/
├── provisioning/     # Cloud provider integrations
├── monitoring/       # Metrics & health checks
├── api/             # Core management API
└── cmd/             # Core API server entrypoint
```

### Commercial Hosting Service (`/hosting-service/`)
```
hosting-service/
├── marketing-site/  # Landing, pricing, testimonials
├── onboarding/      # Signup, billing setup, AI config
├── dashboard/       # Customer dashboard and analytics
├── billing/         # Stripe integration
└── support/         # Ticket system
```

## Service Components

### Marketing Site
- Landing page with feature highlights
- Pricing tiers and comparisons
- Customer testimonials
- Case studies and integrations

### Onboarding App
- User registration wizard
- Payment setup and validation
- AI configuration interface
- Server provisioning initialization

### Customer Dashboard
- Instance management interface
- Real-time analytics and metrics
- Platform settings and configuration
- Support ticket integration

### Billing System
- Stripe payment processing
- Subscription management
- Invoice generation
- Usage-based billing calculations

### Support System
- Ticket creation and tracking
- Knowledge base integration
- Priority queue management
- Escalation workflows

## Integration Points

1. **Customer Registration** → Stripe Billing Setup
2. **Server Provisioning** → Core API Integration
3. **Real-Time Metrics** → Monitoring System
4. **Support Tickets** → Backend Systems
5. **Analytics** → Aggregated Usage Data

See [hosting-service setup guide](../README.md) for detailed implementation information.
