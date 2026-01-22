# Storage API - System Architecture

This document describes the overall architecture of the Storage API system, including hardware, software components, and data flow.

---

## Overview

The Storage API is a personal/household media management system running on a Raspberry Pi. It provides a REST API for managing users, households, and media items (photos, videos, documents).

---

## Hardware

### Raspberry Pi (Main Server)
- **Model**: Raspberry Pi (ARM64 architecture)
- **Hostname**: `storage-pi`
- **Network**:
  - Local: `192.168.100.50`
  - Tailscale: `100.95.169.64`
- **Role**: Runs the API server, PostgreSQL database, and manages storage

### Storage
| Component | Type | Capacity | Purpose |
|-----------|------|----------|---------|
| SD Card / SSD | Internal | ~32-256GB | OS, API binary, PostgreSQL data |
| HDD Bay (Acasis) | External | 2x 4TB | Media storage (RAID 1 mirrored) |

### Storage Layout

```
Pi Internal Storage:
├── /home/pi/storage-api/        # API binary and config
│   ├── storage-api              # Compiled Go binary
│   └── .env                     # Environment variables
├── /home/pi/scripts/            # Maintenance scripts
│   ├── backup-db.sh             # Automated DB backup script
│   └── backup.log               # Backup log file
├── PostgreSQL Data              # Docker volume (pgdata)
│   └── User data, metadata, references to media

External RAID Storage (/dev/md0):
└── /mnt/storage/                # RAID 1 mount point (3.6TB usable)
    ├── media/
    │   ├── photos/
    │   ├── videos/
    │   ├── documents/
    │   └── .thumbs/
    └── backups/
        └── postgres/            # Automated weekly DB backups
```

---

## Software Components

### 1. Storage API (Go)
- **Location**: `/home/pi/storage-api/storage-api`
- **Port**: `8080`
- **Purpose**: REST API for managing users, households, and media
- **Managed by**: systemd (`storage-api.service`)

### 2. PostgreSQL Database
- **Version**: PostgreSQL 16
- **Container**: `storage-postgres` (Docker)
- **Port**: `5432`
- **Purpose**: Stores user accounts, household data, media metadata
- **Data Location**: Docker volume `pgdata` (on Pi's internal storage)

### 3. RAID Storage
- **Type**: RAID 1 (mirrored) via `mdadm`
- **Device**: `/dev/md0` (2x 4TB HDDs via Acasis bay)
- **Mount Point**: `/mnt/storage`
- **Usable Space**: 3.6TB
- **Filesystem**: ext4
- **Purpose**: Store media files and database backups
- **Redundancy**: Full mirror - data exists on both drives
- **Features**: Write-intent bitmap enabled for fast recovery

### 4. Automated Backups
- **Script**: `/home/pi/scripts/backup-db.sh`
- **Schedule**: Weekly (Sunday 2:00 AM via cron)
- **Retention**: 30 days
- **Storage**: `/mnt/storage/backups/postgres/`
- **Format**: Compressed SQL dumps (`.sql.gz`)

---

## Data Model

### What Lives Where

| Data Type | Storage Location | Notes |
|-----------|------------------|-------|
| User accounts | PostgreSQL | Email, roles, auth info |
| Household info | PostgreSQL | Family/group membership |
| Media metadata | PostgreSQL | File references, tags, dates |
| Actual media files | RAID Storage | Photos, videos, documents |
| API configuration | `.env` file | Database URL, ports |

### Database Schema

```
households
├── id (UUID)
├── name
└── created_at

users
├── id (UUID)
├── household_id (FK → households)
├── external_sub (OAuth provider ID)
├── email
├── role (admin/member)
└── created_at

media_items (planned)
├── id (UUID)
├── household_id (FK → households)
├── uploaded_by (FK → users)
├── file_path (path on RAID storage)
├── file_type
├── file_size
├── metadata (JSONB - EXIF, etc.)
└── created_at
```

---

## Network Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ (Tailscale VPN)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Raspberry Pi                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐  │
│  │ Storage API │───▶│ PostgreSQL  │    │  RAID Storage   │  │
│  │  (Go:8080)  │    │  (5432)     │    │  (/mnt/storage) │  │
│  └─────────────┘    └─────────────┘    └─────────────────┘  │
│         │                  │                    │            │
│         └──────────────────┼────────────────────┘            │
│                            │                                 │
│              Docker Network / Localhost                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ (Tailscale / Local Network)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Clients                                 │
│   • Mobile apps (future)                                     │
│   • Web interface (future)                                   │
│   • Desktop clients                                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Deployment

### CI/CD Pipeline (GitHub Actions)

```
Developer pushes to main
         │
         ▼
┌─────────────────┐
│ GitHub Actions  │
│ 1. Checkout     │
│ 2. Build ARM64  │
│ 3. Tailscale    │
│ 4. SCP binary   │
│ 5. Restart svc  │
└─────────────────┘
         │
         ▼ (via Tailscale)
┌─────────────────┐
│  Raspberry Pi   │
│ • Receives bin  │
│ • Restarts API  │
│ • Health check  │
└─────────────────┘
```

### Manual Deployment

```bash
# On Mac (development machine)
GOOS=linux GOARCH=arm64 go build -o storage-api ./cmd/server
scp -i ~/.ssh/github_actions_tailscale storage-api pi@100.95.169.64:/home/pi/storage-api/
ssh -i ~/.ssh/github_actions_tailscale pi@100.95.169.64 "sudo systemctl restart storage-api"
```

---

## Security

### Network Security
- Tailscale VPN for remote access (no exposed ports)
- Local network access on port 8080
- SSH via Tailscale only (for GitHub Actions)

### Authentication (Planned)
- OAuth/OIDC integration (external_sub field ready)
- JWT tokens for API authentication
- Role-based access (admin/member)

### Data Security
- RAID 1 for data redundancy
- Database credentials in `.env` (not in repo)
- PostgreSQL password authentication

---

## Current State vs Future

### Currently Working
- [x] API server running on Pi
- [x] PostgreSQL database with user/household tables
- [x] GitHub Actions deployment pipeline (builds ARM64, deploys via Tailscale)
- [x] Tailscale connectivity
- [x] Health check endpoints
- [x] RAID 1 storage (3.6TB mirrored)
- [x] Automated weekly database backups

### In Progress
- [ ] Media file upload/download endpoints

### Planned
- [ ] OAuth authentication integration
- [ ] Mobile app
- [ ] Web interface
- [ ] Photo/video auto-backup from phones
- [ ] Thumbnail generation
- [ ] Search functionality
- [ ] Off-site backup (cloud sync)

---

## Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ADDR` | API listen address | `0.0.0.0:8080` |
| `DATABASE_URL` | PostgreSQL connection | `postgres://user:pass@localhost:5432/db` |
| `POSTGRES_USER` | Database username | `storageapp` |
| `POSTGRES_PASSWORD` | Database password | `change_me_now` |
| `POSTGRES_DB` | Database name | `storage_db` |

### Service Configuration

**Systemd Service**: `/etc/systemd/system/storage-api.service`
```ini
[Unit]
Description=Storage API Service
After=network-online.target docker.service

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/storage-api
EnvironmentFile=/home/pi/storage-api/.env
ExecStart=/home/pi/storage-api/storage-api
Restart=always
RestartSec=3s

[Install]
WantedBy=multi-user.target
```

---

## API Endpoints

### Current

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Basic health check |
| GET | `/health/db` | Database connectivity check |
| GET | `/v1/me` | Get current user info |

### Planned

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/media/upload` | Upload media file |
| GET | `/v1/media/:id` | Get media file |
| GET | `/v1/media` | List media items |
| DELETE | `/v1/media/:id` | Delete media item |
| GET | `/v1/household/members` | List household members |

---

## Open Questions

1. **Authentication**: Which OAuth provider to use? (Auth0, Clerk, Google directly?)
2. **Media organization**: By date? By household? By type?
3. **Thumbnails**: Generate on upload or on-demand?
4. **Off-site backup**: Sync to cloud storage? (Local RAID + weekly DB backups are configured)
5. **Mobile app**: Native iOS/Android or cross-platform (React Native/Flutter)?

---

## Maintenance Commands

### RAID Management
```bash
# Check RAID health
cat /proc/mdstat
sudo mdadm --detail /dev/md0

# Check storage usage
df -h /mnt/storage
```

### Database Backups
```bash
# Run manual backup
/home/pi/scripts/backup-db.sh

# List backups
ls -lh /mnt/storage/backups/postgres/

# Check backup log
cat /home/pi/scripts/backup.log

# Restore from backup
gunzip -c /mnt/storage/backups/postgres/storage_db_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i storage-postgres psql -U storageapp -d storage_db
```

### Service Management
```bash
# Check API status
sudo systemctl status storage-api

# View API logs
sudo journalctl -u storage-api -f

# Restart API
sudo systemctl restart storage-api
```

### Viewing Logs Remotely
```bash
# From your Mac via Tailscale
ssh pi@100.95.169.64 "sudo journalctl -u storage-api -n 50"

# Or stream logs live
ssh pi@100.95.169.64 "sudo journalctl -u storage-api -f"
```

---

## Related Documentation

- [LOCAL_SETUP.md](LOCAL_SETUP.md) - Local development setup
- [TAILSCALE_SETUP.md](TAILSCALE_SETUP.md) - Tailscale/GitHub Actions deployment
- [SEED_DATA.md](SEED_DATA.md) - Database seed data reference

---