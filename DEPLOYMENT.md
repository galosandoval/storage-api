# Deployment Setup

This document explains how to set up automated deployments to your Raspberry Pi via GitHub Actions.

## Overview

When code is merged to `main`, GitHub Actions will automatically:
1. SSH into your Raspberry Pi
2. Pull the latest code
3. Build the application
4. Run database migrations
5. Restart the systemd service
6. Verify the deployment

## Prerequisites

- Raspberry Pi with SSH access
- Cloudflare Tunnel configured (for secure access)
- Systemd service configured for storage-api
- Git repository cloned to `/home/pi/storage-api`

## Setup Instructions

### 1. Generate SSH Key for GitHub Actions

On your **local machine** (not the Pi):

```bash
# Generate a dedicated SSH key for GitHub Actions
ssh-keygen -t ed25519 -C "github-actions-storage-api" -f ~/.ssh/github_actions_storage_api
```

### 2. Add Public Key to Pi

Copy the public key to your Pi:

```bash
# Copy public key to clipboard
cat ~/.ssh/github_actions_storage_api.pub

# SSH into your Pi
ssh pi@your-pi-address

# Add the public key to authorized_keys
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "YOUR_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### 3. Configure Cloudflare Tunnel

If using Cloudflare Tunnel for SSH access:

```bash
# On your Pi, install cloudflared
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64.deb
sudo dpkg -i cloudflared.deb

# Login to Cloudflare
cloudflared tunnel login

# Create tunnel
cloudflared tunnel create storage-pi

# Configure tunnel for SSH
cat > ~/.cloudflared/config.yml <<EOF
tunnel: YOUR_TUNNEL_ID
credentials-file: /home/pi/.cloudflared/YOUR_TUNNEL_ID.json

ingress:
  - hostname: ssh.your-domain.com
    service: ssh://localhost:22
  - service: http_status:404
EOF

# Run tunnel as service
sudo cloudflared service install
sudo systemctl start cloudflared
sudo systemctl enable cloudflared
```

### 4. Add GitHub Secrets

Go to your GitHub repository:
**Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Add these secrets:

| Secret Name | Value | Example |
|-------------|-------|---------|
| `PI_HOST` | Your Pi's hostname/IP via Cloudflare Tunnel | `ssh.your-domain.com` or `192.168.1.100` |
| `PI_USER` | SSH username | `pi` |
| `PI_SSH_KEY` | Private SSH key content | (entire content of `~/.ssh/github_actions_storage_api`) |
| `PI_PORT` | SSH port (optional, default 22) | `22` |

**To copy private key:**
```bash
cat ~/.ssh/github_actions_storage_api
# Copy entire output including the header/footer
```

### 5. Configure Sudoers (for systemctl restart)

On your Pi, allow the `pi` user to restart the service without a password:

```bash
sudo visudo -f /etc/sudoers.d/storage-api
```

Add this line:
```
pi ALL=(ALL) NOPASSWD: /bin/systemctl restart storage-api, /bin/systemctl is-active storage-api, /bin/systemctl status storage-api
```

Save and exit.

### 6. Test Deployment

#### Manual Test on Pi:
```bash
cd /home/pi/storage-api
./deploy.sh
```

#### Trigger GitHub Action:
1. Make a change and push to `main`
2. Go to **Actions** tab in GitHub
3. Watch the deployment progress
4. Check the logs for any errors

Or manually trigger:
1. Go to **Actions** tab
2. Select "Deploy to Pi" workflow
3. Click "Run workflow"

## Troubleshooting

### SSH Connection Failed
```bash
# Test SSH connection locally
ssh -i ~/.ssh/github_actions_storage_api pi@your-pi-address

# Check Cloudflare Tunnel status (on Pi)
sudo systemctl status cloudflared

# Check SSH logs (on Pi)
sudo tail -f /var/log/auth.log
```

### Migration Failed
```bash
# Check database connection (on Pi)
cd /home/pi/storage-api
source .env
goose -dir migrations postgres "$DATABASE_URL" status

# Check PostgreSQL logs
docker logs storage-postgres
```

### Service Failed to Start
```bash
# Check service status (on Pi)
sudo systemctl status storage-api

# Check service logs
sudo journalctl -u storage-api -n 50 --no-pager
```

### Permission Denied (systemctl)
```bash
# Verify sudoers configuration (on Pi)
sudo cat /etc/sudoers.d/storage-api

# Test manually
sudo systemctl restart storage-api
```

## Security Best Practices

- ✅ Use dedicated SSH key for GitHub Actions
- ✅ Use Cloudflare Tunnel instead of exposing SSH port
- ✅ Limit sudo permissions to only required commands
- ✅ Rotate SSH keys periodically
- ✅ Use GitHub Environments for additional protection
- ✅ Enable branch protection rules on `main`

## Rollback

If a deployment fails, SSH into your Pi and:

```bash
cd /home/pi/storage-api

# Rollback to previous commit
git log --oneline -5
git reset --hard PREVIOUS_COMMIT_HASH

# Rebuild and restart
go build -o storage-api ./cmd/server
sudo systemctl restart storage-api
```

## Manual Deployment

You can always deploy manually via SSH:

```bash
# From your local machine
ssh pi@your-pi-address "cd /home/pi/storage-api && ./deploy.sh"
```

## Monitoring

After deployment, verify:
- Service is running: `sudo systemctl status storage-api`
- Health check: `curl http://localhost:8080/health/live`
- Database migrations: `cd /home/pi/storage-api && goose -dir migrations postgres "$DATABASE_URL" status`
- API logs: `sudo journalctl -u storage-api -f`
