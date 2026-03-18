# ClawHost Launch Guide 🚀

This guide walks you through launching ClawHost from development to production, covering both the open source core and commercial service.

## Pre-Launch Checklist

### Technical Requirements
- [ ] Go 1.23+ installed
- [ ] PostgreSQL database ready
- [ ] Hetzner Cloud account and API token
- [ ] Domain name configured (for production)
- [ ] SSL certificates ready (Let's Encrypt recommended)
- [ ] Docker installed (for containerized deployment)

### Business Requirements (Commercial Service)
- [ ] Stripe account configured
- [ ] Business registration complete
- [ ] Terms of Service and Privacy Policy
- [ ] Support infrastructure ready
- [ ] Pricing strategy finalized

## Development Setup

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/clawhost.git
cd clawhost

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file:

```bash
# Database Configuration
DATABASE_URL="postgres://user:password@localhost:5432/clawhost_dev?sslmode=disable"

# Hetzner Cloud API
HETZNER_API_TOKEN="your_hetzner_token_here"

# JWT Secret (generate with: openssl rand -base64 32)
JWT_SECRET="your_jwt_secret_here"

# Development settings
PORT="8080"
GIN_MODE="debug"
LOG_LEVEL="debug"

# Commercial service (optional for development)
STRIPE_SECRET_KEY="sk_test_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
STRIPE_PRICE_ID_STARTER="price_..."
STRIPE_PRICE_ID_PRO="price_..."
STRIPE_PRICE_ID_ENTERPRISE="price_..."
```

### 3. Database Setup

```bash
# Create database
createdb clawhost_dev

# Run migrations (auto-migrates on first run)
go run core/cmd/main.go
```

### 4. Run Development Servers

```bash
# Terminal 1: Core API Server
go run core/cmd/main.go

# Terminal 2: Commercial Service (if developing commercial features)
go run hosting-service/cmd/main.go

# Terminal 3: Test the API
curl http://localhost:8080/health
```

## Testing Your Setup

### 1. Basic Deployment Test

```bash
# Set your Hetzner token
export HETZNER_API_TOKEN="your_token"

# Run the basic deployment example
cd examples/basic-deployment
go run main.go
```

### 2. API Testing

```bash
# Health check
curl -X GET http://localhost:8080/health

# Core API endpoints
curl -X GET http://localhost:8080/api/v1/instances/test-123/status
curl -X GET http://localhost:8080/api/v1/instances/test-123/metrics
```

### 3. Run Test Suite

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
make test-integration
```

## Production Deployment

### Option 1: Docker Deployment (Recommended)

#### 1. Build Docker Image

```bash
# Build the image
docker build -t clawhost:latest .

# Or use Docker Compose
docker-compose build
```

#### 2. Production Environment Variables

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  clawhost:
    image: clawhost:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:password@db:5432/clawhost
      - HETZNER_API_TOKEN=${HETZNER_API_TOKEN}
      - JWT_SECRET=${JWT_SECRET}
      - GIN_MODE=release
      - LOG_LEVEL=info
    depends_on:
      - db
      - redis
    restart: unless-stopped

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=clawhost
      - POSTGRES_USER=clawhost
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    restart: unless-stopped

volumes:
  postgres_data:
```

#### 3. Deploy with SSL

```bash
# Create production environment file
echo "DB_PASSWORD=$(openssl rand -base64 32)" > .env.prod
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env.prod
echo "HETZNER_API_TOKEN=your_token" >> .env.prod

# Deploy with SSL proxy (using Traefik)
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### Option 2: Direct Deployment

#### 1. Build Binary

```bash
# Build optimized binary
cd core && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../clawhost ./cmd

# Make executable
chmod +x clawhost
```

#### 2. Create Systemd Service

```bash
# Create service file
sudo tee /etc/systemd/system/clawhost.service > /dev/null <<EOF
[Unit]
Description=ClawHost API Server
After=network.target

[Service]
Type=simple
User=clawhost
WorkingDirectory=/opt/clawhost
ExecStart=/opt/clawhost/clawhost
Restart=always
RestartSec=10

# Environment
Environment=DATABASE_URL=postgres://clawhost:password@localhost:5432/clawhost
Environment=HETZNER_API_TOKEN=your_token
Environment=JWT_SECRET=your_jwt_secret
Environment=GIN_MODE=release
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl enable clawhost
sudo systemctl start clawhost
sudo systemctl status clawhost
```

#### 3. Setup Nginx Reverse Proxy

```nginx
# /etc/nginx/sites-available/clawhost
server {
    listen 80;
    server_name your-domain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:MozTLS:10m;
    ssl_session_tickets off;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    
    # Proxy to ClawHost API
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

```bash
# Enable site and reload nginx
sudo ln -s /etc/nginx/sites-available/clawhost /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Monitoring and Maintenance

### 1. Health Monitoring

```bash
# Create health check script
tee /opt/clawhost/health-check.sh > /dev/null <<'EOF'
#!/bin/bash
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status')
if [ "$HEALTH" != "healthy" ]; then
    echo "ClawHost API is unhealthy: $HEALTH"
    # Send alert (email, Slack, etc.)
    exit 1
fi
echo "ClawHost API is healthy"
EOF

chmod +x /opt/clawhost/health-check.sh

# Add to crontab (check every 5 minutes)
echo "*/5 * * * * /opt/clawhost/health-check.sh" | crontab -
```

### 2. Log Management

```bash
# Setup log rotation
sudo tee /etc/logrotate.d/clawhost > /dev/null <<EOF
/var/log/clawhost/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 clawhost clawhost
    postrotate
        systemctl reload clawhost
    endscript
}
EOF
```

### 3. Database Backups

```bash
# Create backup script
tee /opt/clawhost/backup.sh > /dev/null <<'EOF'
#!/bin/bash
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_DIR="/opt/clawhost/backups"
mkdir -p $BACKUP_DIR

# Database backup
pg_dump clawhost > $BACKUP_DIR/clawhost_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/clawhost_$DATE.sql

# Keep only last 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: clawhost_$DATE.sql.gz"
EOF

chmod +x /opt/clawhost/backup.sh

# Schedule daily backups at 2 AM
echo "0 2 * * * /opt/clawhost/backup.sh" | crontab -
```

## Launch Strategy

### Phase 1: Soft Launch (Community)

1. **GitHub Release**
   ```bash
   # Tag and release v1.0.0
   git tag -a v1.0.0 -m "ClawHost v1.0.0 - Open Source Launch"
   git push origin v1.0.0
   ```

2. **Community Announcement**
   - Post on Hacker News
   - Share in relevant Discord/Slack communities
   - Cross-post on Reddit (r/selfhosted, r/golang)
   - Write launch blog post

3. **Documentation**
   - Ensure all docs are up-to-date
   - Create video tutorials
   - Set up community Discord server

### Phase 2: Public Launch (Commercial Service)

1. **Marketing Website**
   - Landing page with pricing
   - Feature comparison (open source vs. commercial)
   - Customer testimonials
   - Live demo

2. **Content Marketing**
   - Technical blog posts
   - Conference presentations
   - Podcast interviews
   - YouTube tutorials

3. **Partnership Outreach**
   - Cloud provider partnerships
   - Integration with popular tools
   - Reseller programs

## Scaling Considerations

### Horizontal Scaling

```yaml
# kubernetes deployment example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clawhost-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: clawhost-api
  template:
    metadata:
      labels:
        app: clawhost-api
    spec:
      containers:
      - name: clawhost-api
        image: clawhost:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: clawhost-secrets
              key: database-url
```

### Database Scaling

```bash
# Read replicas for query scaling
# Connection pooling with PgBouncer
# Database partitioning for large datasets
```

### Monitoring at Scale

```bash
# Prometheus metrics
# Grafana dashboards
# Alert manager for incidents
# Distributed tracing with Jaeger
```

## Troubleshooting

### Common Issues

1. **Database Connection**
   ```bash
   # Test database connection
   pg_isready -h localhost -p 5432
   ```

2. **Hetzner API Issues**
   ```bash
   # Test API access
   curl -H "Authorization: Bearer $HETZNER_API_TOKEN" https://api.hetzner.cloud/v1/servers
   ```

3. **SSL Certificate**
   ```bash
   # Check certificate expiry
   openssl x509 -in /etc/letsencrypt/live/domain/cert.pem -text -noout
   ```

### Getting Help

- **Documentation**: Check [docs/](../docs/) folder
- **Community**: Join our [Discord server](https://discord.gg/clawhost)
- **Issues**: Report bugs on [GitHub](https://github.com/yourusername/clawhost/issues)
- **Commercial Support**: enterprise@clawhost.com

## Success Metrics

### Technical Metrics
- API response time < 200ms
- 99.9% uptime
- Database query time < 50ms
- Instance provisioning time < 5 minutes

### Business Metrics
- Monthly Active Users (MAU)
- Conversion rate from open source to paid
- Customer acquisition cost (CAC)
- Monthly recurring revenue (MRR)
- Net promoter score (NPS)

---

**🎉 Congratulations on launching ClawHost!**

*Remember: Launch is just the beginning. Focus on user feedback, iterate quickly, and build an amazing community around OpenClaw hosting.*