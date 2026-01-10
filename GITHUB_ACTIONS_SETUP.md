# GitHub Actions Setup for Pi Deployment

This guide walks you through setting up automated deployments from GitHub to your Raspberry Pi.

## Overview

The GitHub Actions workflow (`.github/workflows/deploy.yml`) automatically deploys to your Pi whenever you push to the `main` branch.

## Setup Steps

### 1. Generate SSH Key for GitHub Actions

On your **Mac** (not the Pi), generate a dedicated SSH key:

```bash
# Generate ED25519 key (more secure and smaller)
ssh-keygen -t ed25519 -C "github-actions-storage-api" -f ~/.ssh/github_actions_storage_api

# When prompted:
# - Press Enter for no passphrase (required for automation)
# - Press Enter again to confirm
```

This creates two files:
- `~/.ssh/github_actions_storage_api` (private key - keep secret!)
- `~/.ssh/github_actions_storage_api.pub` (public key - safe to share)

### 2. Add Public Key to Raspberry Pi

Copy the public key to your clipboard:

```bash
cat ~/.ssh/github_actions_storage_api.pub
```

SSH into your Pi and add it to authorized keys:

```bash
# SSH into your Pi
ssh pi@<your-pi-ip-or-hostname>

# On the Pi, add the public key
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "PASTE_YOUR_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
exit
```

### 3. Test SSH Connection

From your Mac, test the new key works:

```bash
ssh -i ~/.ssh/github_actions_storage_api pi@<your-pi-ip-or-hostname>
```

If it connects without asking for a password, you're good! Type `exit` to disconnect.

### 4. Configure Sudoers on Pi (for systemctl)

The deployment script needs to restart the service. Configure passwordless sudo for specific commands:

```bash
# SSH into your Pi
ssh pi@<your-pi-ip-or-hostname>

# Edit sudoers
sudo visudo -f /etc/sudoers.d/storage-api
```

Add this line (replace `pi` with your username if different):

```
pi ALL=(ALL) NOPASSWD: /bin/systemctl restart storage-api, /bin/systemctl is-active storage-api, /bin/systemctl status storage-api
```

Save and exit (`Ctrl+X`, then `Y`, then `Enter` in nano).

Test it works:

```bash
sudo systemctl status storage-api
# Should NOT ask for password
```

### 5. Add GitHub Secrets

Go to your GitHub repository:

**Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Add these four secrets:

#### PI_HOST
- Your Pi's IP address or hostname
- Examples:
  - Direct SSH: `192.168.1.100` or `raspberrypi.local`
  - Cloudflare Tunnel: `ssh.yourdomain.com`

#### PI_USER
- SSH username (typically `pi`)

#### PI_SSH_KEY
- The **entire content** of the private key file

```bash
# Copy private key to clipboard (macOS)
cat ~/.ssh/github_actions_storage_api | pbcopy

# Or just display it to copy manually
cat ~/.ssh/github_actions_storage_api
```

Copy from `-----BEGIN OPENSSH PRIVATE KEY-----` to `-----END OPENSSH PRIVATE KEY-----` (inclusive).

#### PI_PORT
- SSH port (usually `22`)
- If using Cloudflare Tunnel, still use `22` (the tunnel handles routing)

### 6. Verify Workflow File

The workflow file at `.github/workflows/deploy.yml` should already be created. It will:

- Trigger on push to `main` branch
- Connect to your Pi via SSH
- Run the `deploy.sh` script
- Report success/failure

### 7. Test the Deployment

#### Option A: Push to Main

```bash
# Make a test change
echo "# Test deployment" >> README.md

# Commit and push
git add .
git commit -m "Test: Trigger deployment"
git push origin main
```

#### Option B: Manual Trigger

1. Go to your GitHub repository
2. Click **Actions** tab
3. Select "Deploy to Raspberry Pi" workflow
4. Click **Run workflow** dropdown
5. Click **Run workflow** button

### 8. Monitor Deployment

Watch the deployment progress:

1. Go to **Actions** tab in GitHub
2. Click on the running workflow
3. Click on the "Deploy to Pi" job
4. Watch the logs in real-time

### 9. Verify on Pi

After deployment succeeds, SSH into your Pi and verify:

```bash
# Check service status
sudo systemctl status storage-api

# Check health endpoint
curl http://localhost:8080/health
curl http://localhost:8080/health/db

# Check recent logs
sudo journalctl -u storage-api -n 50 --no-pager
```

## Troubleshooting

### SSH Connection Failed

```bash
# Test SSH connection from your Mac
ssh -i ~/.ssh/github_actions_storage_api pi@<your-pi-host>

# Check Pi's SSH logs
sudo tail -f /var/log/auth.log
```

Common issues:
- Wrong host/port in GitHub secrets
- Private key not copied correctly (must include header/footer)
- Firewall blocking connection
- Pi not accessible from internet (if not on same network)

### Permission Denied (systemctl)

```bash
# On Pi, verify sudoers configuration
sudo cat /etc/sudoers.d/storage-api

# Test manually
sudo systemctl restart storage-api
# Should not ask for password
```

### Deployment Script Fails

```bash
# On Pi, test the deploy script manually
cd /home/pi/storage-api
./deploy.sh
```

Check for:
- Git repository is in `/home/pi/storage-api`
- Script has execute permissions (`chmod +x deploy.sh`)
- Database is running
- Systemd service is configured

### Service Not Starting

```bash
# Check service logs on Pi
sudo journalctl -u storage-api -n 100 --no-pager

# Check if service file exists
systemctl cat storage-api

# Manually test the binary
cd /home/pi/storage-api
./storage-api
```

## Security Best Practices

- ✅ Use dedicated SSH key for GitHub Actions (not your personal key)
- ✅ Limit sudo permissions to only required commands
- ✅ Consider using Cloudflare Tunnel instead of exposing SSH port
- ✅ Enable GitHub branch protection on `main` branch
- ✅ Rotate SSH keys periodically
- ✅ Use GitHub Environments for additional approval gates (optional)

## Rollback Procedure

If a deployment breaks something:

```bash
# SSH into Pi
ssh pi@<your-pi-host>

# Navigate to project
cd /home/pi/storage-api

# View recent commits
git log --oneline -10

# Reset to previous commit
git reset --hard <previous-commit-hash>

# Rebuild and restart
go build -o storage-api ./cmd/server
sudo systemctl restart storage-api

# Verify
sudo systemctl status storage-api
```

## Next Steps

- Set up monitoring/alerting for your Pi
- Configure automatic backups for PostgreSQL
- Add health check monitoring (e.g., UptimeRobot)
- Consider adding a test job before deployment
