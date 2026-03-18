# ClawHost Installation Guide (OpenClaw Self-Hosting)

This guide covers installing ClawHost for OpenClaw self-hosting in local development and production environments.

## Open Source Local Install (Fastest)

Use this mode if you want a simple, self-managed setup on your own machine with the basic dashboard.

```bash
# 1) Clone
git clone https://github.com/abhishek-kureriya/clawhost.git
cd clawhost

# 2) Build and run OSS core
make build
./clawhost

# 3) Verify
curl http://localhost:8080/health
# Open: http://localhost:8080/dashboard
```

This flow is fully open source and does not require `hosting-service/`.

## Quick Start

Get ClawHost running in under 5 minutes:

```bash
# 1. Clone the repository
git clone https://github.com/abhishek-kureriya/clawhost.git
cd clawhost

# 2. Set up environment
cp .env.example .env
# Edit .env with your settings

# 3. Install dependencies
go mod download

# 4. Run the API server
go run core/cmd/main.go

# 5. Test the installation
curl http://localhost:8080/health
```

## Prerequisites

### System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go Version**: 1.23 or later
- **Database**: PostgreSQL 13+ (recommended) or SQLite for development
- **Memory**: Minimum 512MB RAM, 2GB+ recommended
- **Storage**: 1GB free space (more for logs and backups)

### External Services

- **Hetzner Cloud Account**: [Sign up here](https://hetzner.com/cloud)
- **Domain Name**: For production deployments
- **SSL Certificate**: Let's Encrypt recommended

## Installation Methods

### Method 1: Docker (Recommended)

Easiest way to run ClawHost with all dependencies:

#### 1. Install Docker

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose
sudo usermod -aG docker $USER
# Log out and back in for group changes

# macOS (using Homebrew)
brew install docker docker-compose

# Or install Docker Desktop from docker.com
```

#### 2. Run with Docker Compose

```bash
# Clone repository
git clone https://github.com/yourusername/clawhost.git
cd clawhost

# Copy environment template
cp .env.example .env
# Edit .env with your Hetzner API token

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f clawhost
```

#### 3. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","service":"clawhost-core","version":"1.0.0"}
```

### Method 2: Binary Installation

For production servers or when you prefer direct installation:

#### 1. Install Go

```bash
# Download and install Go 1.23+
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### 2. Install PostgreSQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql
CREATE DATABASE clawhost;
CREATE USER clawhost WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE clawhost TO clawhost;
\q
```

#### 3. Build and Install ClawHost

```bash
# Clone repository
git clone https://github.com/yourusername/clawhost.git
cd clawhost

# Install dependencies
go mod download

# Build binary
go build -o clawhost ./core/cmd

# Install binary (optional)
sudo mv clawhost /usr/local/bin/

# Create configuration directory
sudo mkdir -p /etc/clawhost
sudo chown $USER:$USER /etc/clawhost
```

#### 4. Configure Environment

Create `/etc/clawhost/config.env`:

```bash
# Database Configuration
DATABASE_URL="postgres://clawhost:your_secure_password@localhost:5432/clawhost?sslmode=disable"

# Hetzner Cloud
HETZNER_API_TOKEN="your_hetzner_api_token"

# Security
JWT_SECRET="$(openssl rand -base64 32)"

# Server Configuration
PORT="8080"
GIN_MODE="release"
LOG_LEVEL="info"

# Optional: Commercial Service
STRIPE_SECRET_KEY="sk_live_..."
STRIPE_WEBHOOK_SECRET="whsec_..."
```

#### 5. Create Systemd Service

```bash
# Create service file
sudo tee /etc/systemd/system/clawhost.service > /dev/null <<'EOF'
[Unit]
Description=ClawHost API Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=clawhost
Group=clawhost
WorkingDirectory=/opt/clawhost
ExecStart=/usr/local/bin/clawhost
EnvironmentFile=/etc/clawhost/config.env
Restart=always
RestartSec=10

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/clawhost

[Install]
WantedBy=multi-user.target
EOF

# Create user and directories
sudo useradd -r -s /bin/false clawhost
sudo mkdir -p /opt/clawhost
sudo chown clawhost:clawhost /opt/clawhost

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable clawhost
sudo systemctl start clawhost

# Check status
sudo systemctl status clawhost
```

### Method 3: Development Setup

For developers who want to contribute or customize ClawHost:

#### 1. Development Dependencies

```bash
# Install Go, Git, and development tools
sudo apt install golang-go git build-essential curl

# Install air for live reloading (optional)
go install github.com/cosmtrek/air@latest

# Install database migration tool
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### 2. Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/yourusername/clawhost.git
cd clawhost

# Install dependencies
go mod download

# Setup pre-commit hooks
cp scripts/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

# Copy development config
cp .env.example .env
```

#### 3. Development Database

```bash
# Option 1: Use Docker for development database
docker run --name clawhost-db \
  -e POSTGRES_PASSWORD=dev_password \
  -e POSTGRES_DB=clawhost_dev \
  -p 5432:5432 \
  -d postgres:15

# Option 2: Use SQLite for simple development
# Just set DATABASE_URL="sqlite:./clawhost.db" in .env
```

#### 4. Run Development Server

```bash
# Run with live reloading
air

# Or run directly
go run core/cmd/main.go

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Configuration Options

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `HETZNER_API_TOKEN` | Yes | - | Hetzner Cloud API token |
| `JWT_SECRET` | Yes | - | JWT signing secret (32+ chars) |
| `PORT` | No | `8080` | API server port |
| `GIN_MODE` | No | `debug` | Gin mode (debug/release) |
| `LOG_LEVEL` | No | `info` | Log level (debug/info/warn/error) |
| `STRIPE_SECRET_KEY` | No | - | Stripe API key (commercial) |
| `STRIPE_WEBHOOK_SECRET` | No | - | Stripe webhook secret |

### Configuration File

Alternatively, use a YAML configuration file:

```yaml
# config.yaml
database:
  url: "postgres://user:pass@localhost/clawhost"
  max_connections: 10
  max_idle: 5

server:
  port: 8080
  mode: "release"
  log_level: "info"
  
hetzner:
  api_token: "your_token_here"
  default_location: "nbg1"
  default_server_type: "cx11"

commercial:
  stripe:
    secret_key: "sk_live_..."
    webhook_secret: "whsec_..."
    price_ids:
      starter: "price_starter"
      pro: "price_pro"
      enterprise: "price_enterprise"
```

## Security Considerations

### 1. API Token Security

```bash
# Never commit API tokens to version control
echo "*.env" >> .gitignore
echo "config.yaml" >> .gitignore

# Use restricted permissions
chmod 600 /etc/clawhost/config.env

# Consider using HashiCorp Vault or similar for production
```

### 2. Database Security

```bash
# Use strong passwords
openssl rand -base64 32

# Enable SSL for database connections
DATABASE_URL="postgres://user:pass@localhost/db?sslmode=require"

# Restrict database access
# Edit /etc/postgresql/13/main/pg_hba.conf
```

### 3. Network Security

```bash
# Use firewall to restrict access
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
# Don't open port 8080 to public if using reverse proxy
```

## Reverse Proxy Setup

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/clawhost
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    
    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    
    location / {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

### Caddy Configuration (Alternative)

```caddyfile
# Caddyfile
your-domain.com {
    reverse_proxy localhost:8080
    
    # Automatic HTTPS
    tls your-email@example.com
    
    # Rate limiting
    rate_limit {
        zone dynamic {
            key {remote_addr}
            events 100
            window 1m
        }
    }
    
    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000;"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
        -Server
    }
}
```

## SSL Certificate Setup

### Let's Encrypt with Certbot

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal (add to crontab)
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

## Monitoring Setup

### Basic Monitoring

```bash
# Create monitoring script
tee /opt/clawhost/monitor.sh > /dev/null <<'EOF'
#!/bin/bash
LOG_FILE="/var/log/clawhost/monitor.log"
echo "$(date): Checking ClawHost health" >> $LOG_FILE

# Health check
if curl -sf http://localhost:8080/health > /dev/null; then
    echo "$(date): ClawHost is healthy" >> $LOG_FILE
else
    echo "$(date): ClawHost health check failed" >> $LOG_FILE
    # Send alert (email, Slack, etc.)
fi
EOF

chmod +x /opt/clawhost/monitor.sh

# Run every 5 minutes
echo "*/5 * * * * /opt/clawhost/monitor.sh" | crontab -
```

## Troubleshooting

### Common Issues

1. **"Database connection failed"**
   ```bash
   # Check PostgreSQL status
   sudo systemctl status postgresql
   
   # Test connection
   psql $DATABASE_URL -c "SELECT version();"
   ```

2. **"Hetzner API authentication failed"**
   ```bash
   # Test API token
   curl -H "Authorization: Bearer $HETZNER_API_TOKEN" \
        https://api.hetzner.cloud/v1/servers
   ```

3. **"Port already in use"**
   ```bash
   # Find process using port
   sudo lsof -i :8080
   
   # Change port in configuration
   export PORT=8081
   ```

4. **"Permission denied"**
   ```bash
   # Check file permissions
   ls -la /etc/clawhost/
   
   # Fix permissions
   sudo chown clawhost:clawhost /etc/clawhost/config.env
   sudo chmod 600 /etc/clawhost/config.env
   ```

### Getting Help

- **Documentation**: [GitHub Wiki](https://github.com/yourusername/clawhost/wiki)
- **Community**: [Discord Server](https://discord.gg/clawhost)
- **Issues**: [GitHub Issues](https://github.com/yourusername/clawhost/issues)
- **Support**: GitHub Issues / GitHub Discussions

## Next Steps

After installation:

1. **[Launch Guide](launch-guide.md)** - Deploy your first instance
2. **[API Reference](core-api.md)** - Learn the API endpoints
3. **[Examples](../examples/)** - Try the example deployments
4. **[Contributing](../CONTRIBUTING.md)** - Contribute to the project

---

**🎉 Installation Complete!**

*Your ClawHost installation is ready. Start deploying OpenClaw instances and join our community of developers building the future of AI hosting.*