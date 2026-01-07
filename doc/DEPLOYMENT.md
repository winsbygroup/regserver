# Deployment Guide

This guide covers deploying `regserver` to a [DigitalOcean Droplet](https://www.digitalocean.com/products/droplets) 
Ubuntu Server with [Caddy](https://caddyserver.com/) as a reverse proxy. Any hosting service such as [Hetzner](https://www.hetzner.com/),
[Kamatera](https://www.kamatera.com/products/cloud-servers/), [Linode](https://www.linode.com/products/essential-compute/)
or [Vultr](https://www.vultr.com/) would work as well. Resources needed are tiny with the smallest option usually sufficient
(1vCPU, 1GB RAM, 20GB disk).

## Prerequisites

- A DigitalOcean Droplet (Ubuntu Server LTS recommended). Passkey is recommended but requires additional setup.
- A domain name pointing to your Droplet's IP
- SSH access to the Droplet

## Overview

The deployment consists of:

| Component | Role |
|-----------|------|
| **regserver** | Go application running on port 8080 |
| **Caddy** | Reverse proxy with automatic HTTPS |
| **systemd** | Process manager for both services |

```
Internet → Caddy (:443) → regserver (:8080) → SQLite
              ↓
        Auto HTTPS via Let's Encrypt
```

---

## Step 1: Build the Linux Binary

This project uses `mattn/go-sqlite3` which requires CGO. You must build on a Linux environment (cross-compilation from Windows won't work).

### Windows (using WSL)

```bash
# Open WSL and navigate to project
cd /mnt/c/Users/dougw/Source/Products/regserver

# Build
task build
```

### macOS/Linux

```bash
go build -o ./dist/regserver ./cmd/regserver
```

This creates a Linux binary at `./dist/regserver`.

---

## Step 2: Configure DigitalOcean Firewall and Connect

### Create Cloud Firewall (using control panel)

1. In the DigitalOcean control panel, go to **Networking** → **Firewalls** → **Create Firewall**
2. Name it (e.g., `regserver-firewall`)
3. Configure **Inbound Rules**:

| Type | Protocol | Port Range | Sources |
|------|----------|------------|---------|
| SSH | TCP | 22 | All IPv4 |
| HTTP | TCP | 80 | All IPv4 |
| HTTPS | TCP | 443 | All IPv4 |

4. Leave **Outbound Rules** as default (allow all)
5. Under **Apply to Droplets**, select your droplet
6. Click **Create Firewall**

> **Tip:** For better security, restrict SSH sources to your IP address instead of "All IPv4".

### Connect to Droplet via `ssh`

```bash
ssh root@your-droplet-ip
```

---

## Step 3: Install Caddy

```bash
sudo apt update
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list

sudo apt update
sudo apt install caddy
```

---

## Step 4: Create User and Directories

```bash
# Create a dedicated user (no login shell for security)
sudo useradd --system --no-create-home --shell /usr/sbin/nologin regserver

# Create application directory
sudo mkdir -p /opt/regserver

# Create config directory for secrets
sudo mkdir -p /etc/regserver

# Create log directory for Caddy
sudo mkdir -p /var/log/caddy
sudo chown caddy:caddy /var/log/caddy
```

---

## Step 5: Upload Files

From your **local machine** (open a new terminal):

```bash
# Upload the binary
scp ./dist/regserver root@your-droplet-ip:/opt/regserver/

# Upload the Caddyfile (edit deploy/Caddyfile first to set your domain)
scp ./deploy/Caddyfile root@your-droplet-ip:/etc/caddy/Caddyfile

# Upload the systemd service file
scp ./deploy/regserver.service root@your-droplet-ip:/etc/systemd/system/
```

### Create config.yaml on Droplet

The server requires a config file to set the correct port. On the Droplet:

```bash
sudo nano /opt/regserver/config.yaml
```

Add:

```yaml
addr: ":8080"
db_path: "/opt/regserver/registrations.db"
read_timeout: 5s
write_timeout: 10s
idle_timeout: 120s
```

Set ownership:

```bash
sudo chown regserver:regserver /opt/regserver/config.yaml
```

> **Note:** Without this config, the server will start on a random port and Caddy won't be able to proxy to it.

---

## Step 6: Create or Upload Database

You need a SQLite database for the application. First, install sqlite3 command:

```bash
sudo apt install sqlite3
```

Then, choose one of these options:

### Option A: Create a Fresh Database

**Do nothing.** The server will create a new database if one is not found using the env or config settings.

### Option B: Upload Existing Database

If you have production data:

```bash
# Back up first! (on local machine)
task db:backup

# Upload to Droplet
scp ./testdata/registrations.db root@your-droplet-ip:/opt/regserver/
```

### Set Database Permissions (on Droplet)

```bash
sudo chown regserver:regserver /opt/regserver/registrations.db
sudo chmod 644 /opt/regserver/registrations.db
```

> **Important:** The database contains all your customer and license data. Always back up before deploying changes. See the [Backup](#backup) section for automated backup strategies.

---

## Step 7: Configure Environment Variables (for systemd)

On the Droplet, create the environment file (or copy from the .env.example):

```bash
sudo nano /etc/regserver/env
```

Add the following (DB_PATH only if you want to override config.yaml):

```
ADMIN_API_KEY=your-secret-api-key-here
REGISTRATION_SECRET=your-secret-registration-key
DB_PATH=/opt/regserver/registrations.db
```

Secure the file:

```bash
sudo chmod 600 /etc/regserver/env
sudo chown root:root /etc/regserver/env
```

---

## Step 8: Set Permissions

```bash
# Make binary executable
sudo chmod +x /opt/regserver/regserver

# Set ownership of application directory
sudo chown -R regserver:regserver /opt/regserver
```

---

## Step 9: Configure DNS

In your domain registrar or DigitalOcean DNS settings, add A records:

| Type | Name | Value |
|------|------|-------|
| A | @ | your-droplet-ip |
| A | www | your-droplet-ip |

To use a subdomain (like reg.example.com) just include one A record for the subdomain.

Wait for DNS propagation (can take a few minutes to hours).

Verify with:

```bash
dig +short yourdomain.com
```
---

## Step 10: Update the Caddyfile

Edit `/etc/caddy/Caddyfile` to use your domain:

```bash
sudo nano /etc/caddy/Caddyfile
```

The file should look like this (replace `reg.winsbygroup.com` with your domain):

```
yourdomain.com {
    encode gzip

    reverse_proxy localhost:8080

    log {
        output file /var/log/caddy/regserver.log {
            roll_size 10mb
            roll_keep 10
            roll_keep_for 720h
        }
        format json
    }
}
```

Caddy automatically provisions HTTPS certificates via Let's Encrypt.

---

## Step 11: Start the Services

```bash
# Reload systemd to pick up the new service file
sudo systemctl daemon-reload

# Enable services to start on boot
sudo systemctl enable regserver
sudo systemctl enable caddy

# Start the services
sudo systemctl start regserver
sudo systemctl start caddy
```

---

## Step 12: Verify Deployment

```bash
# Check regserver status
sudo systemctl status regserver

# Check Caddy status
sudo systemctl status caddy

# Test locally
curl http://localhost:8080/livez

# Test health endpoint
curl http://localhost:8080/readyz
```

Visit `https://yourdomain.com/web/` in your browser. You should see the login page.

---

## Useful Commands

### Service Management

| Command | Description |
|---------|-------------|
| `sudo systemctl start regserver` | Start the app |
| `sudo systemctl stop regserver` | Stop the app |
| `sudo systemctl restart regserver` | Restart the app |
| `sudo systemctl status regserver` | Check status |

**Note:** the restart will also reload env vars from `/etc/regserver/env`

### Logs

| Command | Description |
|---------|-------------|
| `sudo journalctl -u regserver -f` | Follow regserver logs |
| `sudo journalctl -u caddy -f` | Follow Caddy logs |
| `sudo journalctl -u regserver -n 50` | Last 50 log lines |
| `sudo journalctl -u regserver --since "1 hour ago"` | Logs from last hour |

### Caddy

| Command | Description |
|---------|-------------|
| `sudo systemctl restart caddy` | Restart Caddy |
| `caddy validate --config /etc/caddy/Caddyfile` | Validate config |
| `sudo caddy reload --config /etc/caddy/Caddyfile` | Reload without restart |

---

## Updating the Application

### SSH Configuration

The deploy script uses an SSH config alias instead of a hardcoded IP address. Configure your SSH client before deploying.

**Option A: SSH Config Alias (recommended)**

Add to `~/.ssh/config`:

```
Host do-regserver
    HostName your-droplet-ip
    User root
    IdentityFile ~/.ssh/id_ed25519
```

Then `task deploy` will work automatically.

**Option B: Environment Variable**

Override the SSH target per-command:

```bash
REGSERVER_HOST=root@your-server-ip task deploy
```

### Quick Deploy (recommended)

From WSL or Linux/macOS:

```bash
cd /mnt/c/Users/username/Source/regserver
task deploy
```

This builds, uploads, and restarts the service automatically.

### Manual Deploy

If you prefer to do it step by step:

#### 1. Build New Binary

```bash
# Windows: Open WSL first
cd /mnt/c/Users/username/Source/regserver
go build -o ./dist/regserver ./cmd/regserver

# macOS/Linux
go build -o ./dist/regserver ./cmd/regserver
```

#### 2. Upload New Binary

```bash
scp ./dist/regserver root@your-droplet-ip:/opt/regserver/regserver.new
```

#### 3. Deploy on Droplet

```bash
sudo systemctl stop regserver
sudo mv /opt/regserver/regserver.new /opt/regserver/regserver
sudo chmod +x /opt/regserver/regserver
sudo chown regserver:regserver /opt/regserver/regserver
sudo systemctl start regserver
```

#### 4. Verify

```bash
sudo systemctl status regserver
curl http://localhost:8080/livez
```

---

## Troubleshooting

### regserver won't start

```bash
# Check the logs
sudo journalctl -u regserver -n 100

# Common issues:
# - Missing ADMIN_API_KEY: Check /etc/regserver/env exists and has the key
# - Permission denied: Check ownership with ls -la /opt/regserver/
# - Port in use: Check with sudo lsof -i :8080
```

### Caddy won't start

```bash
# Check the logs
sudo journalctl -u caddy -n 100

# Validate config
caddy validate --config /etc/caddy/Caddyfile

# Common issues:
# - DNS not propagated: Caddy can't get HTTPS cert if domain doesn't resolve
# - Port 80/443 in use: Check with sudo lsof -i :80
```

### Can't get HTTPS certificate

```bash
# Ensure ports 80 and 443 are open
sudo ufw allow 80
sudo ufw allow 443

# Check DNS resolves to your server
dig +short yourdomain.com

# Check Caddy logs for Let's Encrypt errors
sudo journalctl -u caddy | grep -i "certificate\|acme\|tls"
```

### Database issues

```bash
# Check database exists and has correct permissions
ls -la /opt/regserver/registrations.db

# Should be owned by regserver user
sudo chown regserver:regserver /opt/regserver/registrations.db
```

---

## Security Checklist

- [ ] `ADMIN_API_KEY` is set and secure (use `task gen:apikey`)
- [ ] `/etc/regserver/env` has permissions `600` (only root can read)
- [ ] Firewall allows only ports 22, 80, 443
- [ ] regserver runs as non-root user
- [ ] Database file is not world-readable

---

## Backup

> **Critical:** The SQLite database contains all customer registrations, licenses, and machine activations. **Data loss is unrecoverable without backups.** Set up automated backups before going to production.

### Manual Backup (one-time)

```bash
# On the Droplet - create a backup
sqlite3 /opt/regserver/registrations.db ".backup /tmp/registrations_backup.db"

-or-

sqlite3 /opt/regserver/registrations.db "VACUUM INTO '/tmp/temp.db';"
sqlite3 /tmp/temp.db .dump > /tmp/registrations_backup.sql
rm /tmp/temp.db

# Download to local machine
scp root@your-droplet-ip:/tmp/registrations_backup.db ./backups/
```

### Automated Daily Backups (recommended)

#### 1. Create backup script

```bash
sudo nano /opt/regserver/backup.sh
```

Add this content:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/opt/regserver/backups"
DB_PATH="/opt/regserver/registrations.db"
DATE=$(date +%Y%m%d_%H%M%S)
KEEP_DAYS=30

# Create backup directory if needed
mkdir -p "$BACKUP_DIR"

# Create backup (safe for running database)
sqlite3 "$DB_PATH" ".backup $BACKUP_DIR/registrations_$DATE.db"

# Compress the backup
gzip "$BACKUP_DIR/registrations_$DATE.db"

# Delete backups older than KEEP_DAYS
find "$BACKUP_DIR" -name "registrations_*.db.gz" -mtime +$KEEP_DAYS -delete

echo "Backup complete: registrations_$DATE.db.gz"
```

#### 2. Make it executable

```bash
sudo chmod +x /opt/regserver/backup.sh
sudo chown regserver:regserver /opt/regserver/backup.sh
```

#### 3. Create backup directory

```bash
sudo mkdir -p /opt/regserver/backups
sudo chown regserver:regserver /opt/regserver/backups
```

#### 4. Schedule with cron

```bash
sudo crontab -u regserver -e
```

Add this line for daily backups at 2 AM:

```
0 2 * * * /opt/regserver/backup.sh >> /opt/regserver/backups/backup.log 2>&1
```

### Off-site Backups with Backblaze B2 (Recommended)

Backblaze B2 is the cheapest cloud storage option ($0.005/GB/month) and is designed for backups. For a small SQLite database, this costs essentially nothing.

#### 1. Create a B2 Bucket

1. Sign up at [backblaze.com](https://www.backblaze.com/b2/cloud-storage.html)
2. Go to **B2 Cloud Storage** → **Buckets** → **Create a Bucket**
   - Name: `yourcompany-regserver-backups`
   - Privacy: **Private**
3. Go to **App Keys** → **Add a New Application Key**
   - Name: `regserver-backup`
   - Allow access to: Select your bucket
   - Save the `keyID` and `applicationKey` (shown only once!)

#### 2. Install rclone on Droplet

```bash
sudo apt update
sudo apt install rclone
```

#### 3. Configure rclone for B2

```bash
rclone config
```

Follow the prompts:
```
n) New remote
name> b2
Storage> b2
account> YOUR_KEY_ID
key> YOUR_APPLICATION_KEY
hard_delete> false
Edit advanced config> n
y) Yes this is OK
q) Quit config
```

#### 4. Test the connection

```bash
rclone lsd b2:
# Should list your bucket
```

#### 5. Update backup script

Edit the backup script to upload to B2:

```bash
sudo nano /opt/regserver/backup.sh
```

Replace with this updated version:

```bash
#!/bin/bash
set -e

BACKUP_DIR="/opt/regserver/backups"
DB_PATH="/opt/regserver/registrations.db"
DATE=$(date +%Y%m%d_%H%M%S)
KEEP_DAYS=30
B2_BUCKET="yourcompany-regserver-backups"

# Create backup directory if needed
mkdir -p "$BACKUP_DIR"

# Create backup (safe for running database)
sqlite3 "$DB_PATH" ".backup $BACKUP_DIR/registrations_$DATE.db"

# Compress the backup
gzip "$BACKUP_DIR/registrations_$DATE.db"

BACKUP_FILE="$BACKUP_DIR/registrations_$DATE.db.gz"

# Upload to Backblaze B2
if rclone copy "$BACKUP_FILE" "b2:$B2_BUCKET/"; then
    echo "$(date): Uploaded $BACKUP_FILE to B2"
else
    echo "$(date): ERROR - B2 upload failed for $BACKUP_FILE"
fi

# Delete local backups older than KEEP_DAYS (B2 keeps them longer)
find "$BACKUP_DIR" -name "registrations_*.db.gz" -mtime +$KEEP_DAYS -delete

echo "$(date): Backup complete: registrations_$DATE.db.gz"
```

#### 6. Set B2 lifecycle rules (optional)

In the B2 web console, you can set lifecycle rules to automatically delete old backups:

1. Go to your bucket → **Lifecycle Settings**
2. Set "Keep only the last version" for files older than 90 days

This keeps 30 days locally + 90 days in B2 for extra safety.

### Alternative Off-site Options

**DigitalOcean Spaces** (if you prefer staying in DO ecosystem):

```bash
# Install s3cmd
sudo apt install s3cmd
s3cmd --configure  # Enter your Spaces credentials

# Add to backup.sh:
s3cmd put "$BACKUP_FILE" s3://your-bucket/regserver-backups/
```

**Download to local machine** (simple but requires your machine to be on):

```bash
# Add to your local crontab
0 3 * * * scp root@your-droplet-ip:/opt/regserver/backups/$(date +\%Y\%m\%d)*.db.gz ~/backups/regserver/
```

### Restore from Backup

```bash
# Stop the service
sudo systemctl stop regserver

# Decompress backup
gunzip -k /opt/regserver/backups/registrations_20250115_020000.db.gz

# Replace database
sudo mv /opt/regserver/registrations.db /opt/regserver/registrations.db.old
sudo mv /opt/regserver/backups/registrations_20250115_020000.db /opt/regserver/registrations.db
sudo chown regserver:regserver /opt/regserver/registrations.db

# Restart
sudo systemctl start regserver
```
