package provisioning

import "fmt"

// OpenClawConfig represents OpenClaw instance configuration
type OpenClawConfig struct {
	LLMProvider       string            `json:"llm_provider"`
	LLMModel          string            `json:"llm_model"`
	PersonalityPrompt string            `json:"personality_prompt"`
	BusinessKnowledge string            `json:"business_knowledge"`
	EnvironmentVars   map[string]string `json:"environment_vars"`
}

// GenerateCloudInitScript creates a cloud-init script for OpenClaw installation
func GenerateCloudInitScript(config OpenClawConfig) string {
	return fmt.Sprintf(`#cloud-config
users:
  - name: openclaw
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']

packages:
  - docker.io
  - docker-compose
  - nginx
  - curl
  - git
  - htop

runcmd:
  - systemctl enable docker
  - systemctl start docker
  - usermod -aG docker openclaw
  
  # Install OpenClaw
  - git clone https://github.com/openclaw/openclaw.git /opt/openclaw
  - cd /opt/openclaw
  
  # Configure OpenClaw
  - echo "LLM_PROVIDER=%s" >> .env
  - echo "LLM_MODEL=%s" >> .env
  - echo "AI_PERSONALITY=%s" >> .env
  - echo "BUSINESS_KNOWLEDGE=%s" >> .env
  - echo "PORT=3000" >> .env
  
  # Start OpenClaw with Docker
  - docker-compose up -d
  
  # Configure Nginx reverse proxy
  - |
    cat > /etc/nginx/sites-available/openclaw <<EOF
    server {
        listen 80;
        server_name _;
        
        location / {
            proxy_pass http://localhost:3000;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            proxy_cache_bypass \$http_upgrade;
        }
        
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
    EOF
  
  # Enable Nginx site
  - ln -s /etc/nginx/sites-available/openclaw /etc/nginx/sites-enabled/
  - rm -f /etc/nginx/sites-enabled/default
  - systemctl reload nginx
  
  # Final setup
  - chown -R openclaw:openclaw /opt/openclaw
  - systemctl enable nginx
  - systemctl start nginx
`, config.LLMProvider, config.LLMModel, config.PersonalityPrompt, config.BusinessKnowledge)
}

// GetDefaultConfig returns a default OpenClaw configuration
func GetDefaultConfig() OpenClawConfig {
	return OpenClawConfig{
		LLMProvider:       "openai",
		LLMModel:          "gpt-3.5-turbo",
		PersonalityPrompt: "You are a helpful AI assistant.",
		BusinessKnowledge: "",
		EnvironmentVars:   make(map[string]string),
	}
}
