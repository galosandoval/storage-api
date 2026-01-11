# Tailscale Setup for GitHub Actions Deployment

This guide shows you how to set up Tailscale for automated deployments from GitHub Actions to your Raspberry Pi.

## Why Tailscale?

- ✅ Simple setup (5-10 minutes)
- ✅ No browser authentication needed for CI/CD
- ✅ Secure mesh networking with encrypted connections
- ✅ Free for personal use (up to 100 devices)
- ✅ Works seamlessly with GitHub Actions

## Step 1: Install Tailscale on Your Raspberry Pi

SSH to your Pi using whatever method currently works (local network, etc.):

```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Start Tailscale
sudo tailscale up
```

You'll see output like:
```
To authenticate, visit:
  https://login.tailscale.com/a/xxxxxxxxxxxxx
```

Open that URL in your browser and log in with:
- Google
- GitHub
- Microsoft
- Or create a Tailscale account

Approve your Raspberry Pi to join your network.

## Step 2: Get Your Pi's Tailscale IP

After authentication, get your Pi's Tailscale IP:

```bash
tailscale ip -4
```

Example output: `100.64.1.2`

**Save this IP!** You'll need it for GitHub Actions.

## Step 3: (Optional) Install Tailscale on Your Mac

For easier local development:

```bash
# Option 1: Download from website
# Visit https://tailscale.com/download/mac

# Option 2: Install via Homebrew
brew install tailscale
```

Then authenticate by clicking the Tailscale icon in your menu bar.

## Step 4: Test SSH Connection

From your Mac (after installing Tailscale):

```bash
ssh pi@100.64.1.2  # Use your actual Tailscale IP
```

It should connect without any tunnels or proxy commands! 🎉

## Step 5: Generate SSH Key for GitHub Actions

```bash
# Generate a dedicated SSH key for GitHub Actions
ssh-keygen -t ed25519 -C "github-actions-tailscale" -f ~/.ssh/github_actions_tailscale

# Press Enter twice (no passphrase for automation)
```

## Step 6: Add Public Key to Your Pi

Copy the public key:

```bash
cat ~/.ssh/github_actions_tailscale.pub
```

SSH to your Pi and add it:

```bash
ssh pi@100.64.1.2  # your Tailscale IP

# On the Pi:
mkdir -p ~/.ssh
chmod 700 ~/.ssh
nano ~/.ssh/authorized_keys
# Paste the public key on a new line
# Save with Ctrl+X, Y, Enter

chmod 600 ~/.ssh/authorized_keys
exit
```

Test the key works:

```bash
ssh -i ~/.ssh/github_actions_tailscale pi@100.64.1.2
```

## Step 7: Create Tailscale OAuth Credentials

GitHub Actions needs OAuth credentials to connect to your Tailscale network:

1. Go to https://login.tailscale.com/admin/settings/oauth
2. Click **Generate OAuth client**
3. Give it a name: `GitHub Actions`
4. Under **Tags**, add: `tag:ci`
5. Click **Generate client**
6. **Save both values immediately** (you'll only see the secret once!):
   - **Client ID** (starts with `k...`)
   - **Client Secret** (starts with `tskey-client-...`)

## Step 8: Configure ACL for CI Tag

Allow GitHub Actions to access your Pi:

1. Go to https://login.tailscale.com/admin/acls
2. Add this ACL (or modify existing):

```json
{
  "tagOwners": {
    "tag:ci": ["autogroup:admin"]
  },
  "acls": [
    {
      "action": "accept",
      "src": ["tag:ci"],
      "dst": ["*:*"]
    }
  ]
}
```

3. Click **Save**

This allows devices tagged with `tag:ci` (GitHub Actions) to access all your Tailscale devices.

## Step 9: Add GitHub Secrets

Go to your GitHub repository:
**Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Add these **4 secrets**:

### `TS_OAUTH_CLIENT_ID`
The OAuth Client ID from Step 7 (starts with `k...`)

### `TS_OAUTH_SECRET`
The OAuth Client Secret from Step 7 (starts with `tskey-client-...`)

### `PI_TAILSCALE_IP`
Your Pi's Tailscale IP from Step 2 (e.g., `100.64.1.2`)

### `PI_USER`
Your Pi's SSH username (usually `pi`)

### `PI_SSH_KEY`
Your private SSH key:

```bash
cat ~/.ssh/github_actions_tailscale
```

Copy the entire output including `-----BEGIN OPENSSH PRIVATE KEY-----` and `-----END OPENSSH PRIVATE KEY-----`

## Step 10: Configure Sudoers on Pi

Allow the Pi user to restart the service without password:

```bash
# SSH to your Pi
ssh pi@100.64.1.2

# Edit sudoers
sudo visudo -f /etc/sudoers.d/storage-api
```

Add this line:
```
pi ALL=(ALL) NOPASSWD: /bin/systemctl restart storage-api, /bin/systemctl is-active storage-api, /bin/systemctl status storage-api
```

Save and exit. Test it:

```bash
sudo systemctl status storage-api
# Should NOT ask for password
```

## Step 11: Push and Deploy!

```bash
# Commit the workflow changes
git add .github/workflows/deploy.yml
git commit -m "Switch to Tailscale for GitHub Actions deployment"

# Push to main to trigger deployment
git push origin main
```

Watch the deployment:
1. Go to your GitHub repository
2. Click **Actions** tab
3. Watch the "Deploy to Raspberry Pi" workflow run

## Troubleshooting

### Tailscale Connection Failed

```bash
# On your Pi, check Tailscale status
sudo tailscale status

# Restart Tailscale if needed
sudo systemctl restart tailscaled
```

### SSH Permission Denied

```bash
# Verify the public key is on the Pi
ssh pi@100.64.1.2
cat ~/.ssh/authorized_keys
# Should contain the github_actions_tailscale.pub key
```

### OAuth Error in GitHub Actions

Double-check:
- `TS_OAUTH_CLIENT_ID` and `TS_OAUTH_SECRET` are correct
- ACL includes `tag:ci` with proper permissions
- OAuth client wasn't deleted or expired

### Can't Reach Pi from GitHub Actions

```bash
# On your Pi, verify it's connected to Tailscale
tailscale status

# Check if Pi is advertising routes
tailscale status --json | grep -i online
```

## Security Notes

- ✅ Tailscale uses WireGuard encryption (industry standard)
- ✅ OAuth credentials are scoped to specific tags (`tag:ci`)
- ✅ SSH keys are dedicated to GitHub Actions (not your personal keys)
- ✅ Connections are peer-to-peer when possible, or routed through Tailscale's DERP servers
- ✅ No ports exposed to the public internet

## Useful Commands

```bash
# Check Tailscale status on Pi
sudo tailscale status

# See your Tailscale IP
tailscale ip -4

# See all devices on your network
tailscale status

# Disconnect from Tailscale (if needed)
sudo tailscale down

# Reconnect
sudo tailscale up

# View logs
sudo journalctl -u tailscaled -f
```

## Benefits Over Cloudflare Access

| Feature | Tailscale | Cloudflare Access |
|---------|-----------|-------------------|
| Setup Time | 5-10 min | 1-2 hours |
| Browser Auth for CI | ❌ Not needed | ✅ Required (hard to automate) |
| Device Management | Simple | Complex (WARP, certificates) |
| Free Tier | 100 devices | Limited |
| GitHub Actions Support | Native | Workarounds needed |

## Next Steps

- Consider setting up [Tailscale SSH](https://tailscale.com/kb/1193/tailscale-ssh) for even simpler key management
- Enable [MagicDNS](https://tailscale.com/kb/1081/magicdns) to use hostnames instead of IPs
- Set up [exit nodes](https://tailscale.com/kb/1103/exit-nodes) if you want to route all Pi traffic through another device

---

Need help? Check the [Tailscale documentation](https://tailscale.com/kb/) or [community forums](https://forum.tailscale.com/).
