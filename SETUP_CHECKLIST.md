# GitHub Actions Deployment Setup Checklist

Follow these steps to enable automated deployments to your Raspberry Pi.

## Prerequisites
- [ ] Raspberry Pi is accessible (you're doing this on it! ✅)
- [ ] Git repository: `github.com/galosandoval/storage-api`
- [ ] Cloudflare account for tunneling

---

## Step 1: Generate SSH Key (On Local Machine)

```bash
# Generate dedicated SSH key for GitHub Actions
ssh-keygen -t ed25519 -C "github-actions-storage-api" -f ~/.ssh/github_actions_storage_api

# Display public key (copy this)
cat ~/.ssh/github_actions_storage_api.pub
```

**Action:** Copy the public key output

---

## Step 2: Add Public Key to Pi

```bash
# On your Pi (you're already here!)
echo "PASTE_YOUR_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys

# Verify it was added
tail -1 ~/.ssh/authorized_keys
```

---

## Step 3: Configure Sudoers (On Pi)

```bash
# Allow systemctl restart without password
sudo tee /etc/sudoers.d/storage-api <<EOF
pi ALL=(ALL) NOPASSWD: /bin/systemctl restart storage-api, /bin/systemctl is-active storage-api, /bin/systemctl status storage-api
EOF

# Test it works (should not ask for password)
sudo systemctl status storage-api
```

---

## Step 4: Setup Cloudflare Tunnel (On Pi)

### Install cloudflared:
```bash
# Download for ARM64
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64.deb

# Install
sudo dpkg -i cloudflared.deb

# Verify
cloudflared --version
```

### Login and create tunnel:
```bash
# Login (opens browser)
cloudflared tunnel login

# Create tunnel
cloudflared tunnel create storage-pi

# Note the tunnel ID shown in output
```

### Configure tunnel:
```bash
# Create config (replace YOUR_TUNNEL_ID with actual ID)
mkdir -p ~/.cloudflared
cat > ~/.cloudflared/config.yml <<EOF
tunnel: YOUR_TUNNEL_ID
credentials-file: /home/pi/.cloudflared/YOUR_TUNNEL_ID.json

ingress:
  - hostname: ssh.your-domain.com
    service: ssh://localhost:22
  - service: http_status:404
EOF
```

### Create DNS record:
Go to Cloudflare dashboard:
1. Select your domain
2. DNS → Records → Add record
3. Type: CNAME
4. Name: `ssh` (or whatever you used above)
5. Target: `YOUR_TUNNEL_ID.cfargotunnel.com`
6. Proxy status: Proxied

### Start tunnel as service:
```bash
# Install service
sudo cloudflared service install

# Start it
sudo systemctl start cloudflared
sudo systemctl enable cloudflared

# Check status
sudo systemctl status cloudflared
```

---

## Step 5: Test SSH Connection (On Local Machine)

```bash
# Test connection through Cloudflare Tunnel
ssh -i ~/.ssh/github_actions_storage_api pi@ssh.your-domain.com

# If successful, you should be connected to your Pi!
```

---

## Step 6: Add GitHub Secrets

Go to: https://github.com/galosandoval/storage-api/settings/secrets/actions

### Add these secrets:

#### 1. `PI_HOST`
```
Value: ssh.your-domain.com
(or your Pi's IP if not using tunnel)
```

#### 2. `PI_USER`
```
Value: pi
```

#### 3. `PI_SSH_KEY`
On your local machine:
```bash
# Copy entire private key
cat ~/.ssh/github_actions_storage_api
```
Copy the ENTIRE output (including `-----BEGIN` and `-----END` lines)

Paste into GitHub secret

#### 4. `PI_PORT` (optional)
```
Value: 22
(only if using non-standard port)
```

---

## Step 7: Test Deployment

### Option A: Manual trigger
1. Go to: https://github.com/galosandoval/storage-api/actions
2. Click "Deploy to Pi" workflow
3. Click "Run workflow" → "Run workflow"
4. Watch the logs

### Option B: Push to main
```bash
# On your local machine
cd /path/to/storage-api
git checkout main
git pull

# Make a test change
echo "# Test deployment" >> README.md
git add README.md
git commit -m "test: trigger deployment"
git push origin main

# Watch GitHub Actions run
```

### Option C: Test locally on Pi first
```bash
# On your Pi
cd /home/pi/storage-api
./deploy.sh

# Should see:
# 🚀 Starting deployment...
# 📦 Pulling latest code...
# 🔨 Building application...
# 🗄️  Running migrations...
# 🔄 Restarting service...
# ✅ Service is running successfully
# 🎉 Deployment complete!
```

---

## Troubleshooting

### SSH connection fails
```bash
# Check Cloudflare tunnel (on Pi)
sudo systemctl status cloudflared
sudo journalctl -u cloudflared -n 50

# Check SSH is running (on Pi)
sudo systemctl status ssh

# Test from local machine
ssh -vvv -i ~/.ssh/github_actions_storage_api pi@ssh.your-domain.com
```

### Deployment fails at migration step
```bash
# Check database (on Pi)
docker ps | grep postgres
docker logs storage-postgres

# Test migration manually
cd /home/pi/storage-api
./migrate.sh
```

### Service fails to restart
```bash
# Check service (on Pi)
sudo systemctl status storage-api
sudo journalctl -u storage-api -n 50

# Check sudoers
sudo cat /etc/sudoers.d/storage-api
```

---

## Verification

After setup, verify everything works:

- [ ] SSH works: `ssh -i ~/.ssh/github_actions_storage_api pi@ssh.your-domain.com`
- [ ] Deploy script works: `cd /home/pi/storage-api && ./deploy.sh`
- [ ] GitHub Actions secrets are set
- [ ] Cloudflare Tunnel is running
- [ ] Push to `main` triggers deployment
- [ ] Service restarts successfully

---

## Security Notes

✅ **Good practices:**
- Dedicated SSH key (not your personal key)
- Cloudflare Tunnel (no exposed ports)
- Limited sudo permissions
- Private repository

🔒 **Keep secure:**
- Never commit `.env` file
- Rotate SSH keys periodically
- Monitor GitHub Actions logs
- Use GitHub branch protection on `main`

---

## Quick Reference

| Task | Command |
|------|---------|
| Test deployment locally | `cd /home/pi/storage-api && ./deploy.sh` |
| Check service status | `sudo systemctl status storage-api` |
| View service logs | `sudo journalctl -u storage-api -f` |
| Check tunnel status | `sudo systemctl status cloudflared` |
| Restart tunnel | `sudo systemctl restart cloudflared` |
| Test SSH | `ssh -i ~/.ssh/github_actions_storage_api pi@ssh.your-domain.com` |

---

**Need help?** See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed documentation.


