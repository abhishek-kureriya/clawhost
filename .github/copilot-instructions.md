# ClawHost - Managed OpenClaw AI Hosting Platform

## Project Overview
ClawHost is a SaaS platform providing managed OpenClaw AI hosting for businesses. The platform provisions dedicated Hetzner Cloud servers for customers, handling server management, monitoring, backups, and support.

## Architecture
- **Backend**: Go with Gin framework, GORM ORM, PostgreSQL
- **Frontend**: Next.js 14 with TypeScript and Tailwind CSS
- **Cloud**: Hetzner Cloud API integration
- **Payments**: Stripe
- **Auth**: JWT with middleware
- **Database**: PostgreSQL
- **Deployment**: Docker containers

## Business Model
- Monthly subscriptions: €49-199
- Hetzner Cloud servers: €4-20/month cost
- Customers provide LLM API keys
- Managed hosting service

## Development Status
✅ Project structure created
⬜ Go backend implementation
⬜ Next.js frontend setup
⬜ Hetzner Cloud integration
⬜ Stripe payment processing
⬜ Customer provisioning system