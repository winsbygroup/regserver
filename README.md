# Registration Service (regserver)

[![Live Demo](https://img.shields.io/badge/Demo-live-brightgreen)](https://regserver.onrender.com/web)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-O'Sassy-blue)](https://osaasy.dev/)

A complete software license registration process for **desktop applications**. It provides APIs for client app software 
activation and administrative management of customers, products, licenses, and machine registrations. It also includes 
a complete web interface for administering these resources and viewing license activity and expirations.

## Features

- **Subscription & Perpetual Licenses** - Support for time-limited subscriptions and perpetual licenses with optional maintenance expiration
- **Feature Flags** - Define product features (integer, string, or enum types) with per-customer overrides (e.g. paid subscription levels)
- **License Activation** - Clients activate products using license keys with automatic seat tracking
- **Multi-Machine Support** - Track registrations across multiple machines per license with configurable seat limits
- **Admin REST API** - Full CRUD operations for customers, products, licenses, and registrations
- **Web Admin UI** - Browser-based management with a modern feel (reactive controls with light and dark themes)
- **Offline Registration** - Manual registration workflow for customers without internet access
- **Version Tracking** - Track installed versions and notify clients of available updates (with download links)
- **Registration Tracking** - View machine registrations, installed product versions in use and export expirations to a csv.
- **Client Integration** - Full documentation to implement the client-side activation and validation process (sample code in C#, Delphi and Go)
- **Simple Deployment** - One executable requiring very small resources (full documentation with example for $7/mo DigitalOcean droplet)

## Tech Stack

- **Go 1.24+** with Echo v4 HTTP router
- **SQLite** with WAL mode
- **sqlx** database extensions (https://github.com/jmoiron/sqlx)
- **Templ** + **HTMX** for web UI (https://templ.guide/) (https://htmx.org/)
- **DaisyUI/Tailwind CSS** via CDN (no build step required) (https://daisyui.com/)
- **Tom Select** for searchable dropdowns (via CDN) (https://tom-select.js.org/)
- **Task** for `make`-like task management (https://taskfile.dev/)

## Quick Start (from source)

```bash
# Install go
https://go.dev/doc/install

# Install go dependencies
go mod tidy
go install github.com/a-h/templ/cmd/templ@latest

# Build
task build
# Or: templ generate && go build -o ./dist/regserver ./cmd/regserver

# Run (with sample data loaded)
./dist/regserver -demo
```

Server starts on `:8080`.

---

# 1. Client Software Registration API

**Purpose:** Allows client software to activate licenses, validate registrations, and check for updates.

Each customer receives a unique **license key** (GUID) when they purchase a product. The client software uses this 
license key to authenticate API requests and activate the product on individual machines.

## Authentication

| Endpoint | Authentication |
|----------|----------------|
| `POST /activate` | `X-License-Key` header required |
| `GET /license/:license_key` | License key in URL (self-authenticating) |
| `PUT /license/:license_key` | License key in URL (self-authenticating) |
| `GET /productver/:product_guid` | Public, no auth required |

**Example activation request:**
```
POST /api/v1/activate HTTP/1.1
Host: localhost:8080
X-License-Key: 287d3e24-af8e-4f45-99e8-a9e9f1ca1a91
```

## Endpoints

Base path: `/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/activate` | Activate a product for a machine |
| GET | `/license/:license_key` | Get license information and availability |
| PUT | `/license/:license_key` | Update installed version for a machine |
| GET | `/productver/:product_guid` | Get product version info (public) |

### POST `/activate`

Activate a product for a machine. Creates a new registration or updates an existing one. It is up to the client
to save this response to local storage on the registered machine (see [Client Implementation](doc/clients/README.md)
documentation).

**Request:**
```json
{
  "machineCode": "5mToXAaMQRRXOG58VT2oRKBgD8c=nWxB5pHxLwJx/LbewudPWXecK3c=",
  "userName": "Joe User"
}
```

**Response:**
```json
{
  "UserName": "Joe User",
  "UserCompany": "Winsby Group LLC",
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "MachineCode": "5mToXAaMQRRXOG58VT2oRKBgD8c=nWxB5pHxLwJx/LbewudPWXecK3c=",
  "ExpirationDate": "2025-12-31",
  "MaintExpirationDate": "2025-12-31",
  "MaxProductVersion": "",
  "LatestVersion": "5.5.1",
  "LicenseKey": "287d3e24-af8e-4f45-99e8-a9e9f1aa1a91",
  "RegistrationHash": "ogwa5eQEaFjN/28bsapgee3cyH0=",
  "Features": {
    "Legacy": true,
    "PartTypes": 999999999,
    "Structured": true
  }
}
```

### GET `/license/:license_key`

Get license information including available license count. This endpoint is useful for client software to check license 
availability before attempting activation.

**Response:**
```json
{
  "CustomerName": "Acme Corp",
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "ProductName": "AceMapper",
  "LicenseCount": 5,
  "LicensesAvailable": 3,
  "ExpirationDate": "2025-12-31",
  "MaintExpirationDate": "2025-12-31",
  "MaxProductVersion": "4.5",
  "LatestVersion": "5.5.1",
  "Features": {
    "Legacy": "True",
    "PartTypes": "999999999",
    "Structured": "True"
  }
}
```

| Field | Description |
|-------|-------------|
| `CustomerName` | Name of the customer who owns this license |
| `LicenseCount` | Total number of licenses purchased |
| `LicensesAvailable` | Remaining licenses (only counts non-expired registrations as "in use") |
| `MaxProductVersion` | Maximum version allowed (empty = no restriction) |
| `LatestVersion` | Latest available product version |

### PUT `/license/:license_key`

Update the installed version for a registered machine. This endpoint allows client software to report which version is 
currently installed, useful for tracking deployments and prompting updates.

**Request:**
```json
{
  "machineCode": "5mToXAaMQRRXOG58VT2oRKBgD8c=nWxB5pHxLwJx/LbewudPWXecK3c=",
  "installedVersion": "5.5.0"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `machineCode` | Yes | The machine code identifying the registered machine |
| `installedVersion` | No | The version currently installed on the machine |

**Response:** Same as GET `/license/:license_key`

**Errors:**
- `404 Not Found` - License key not found, or machine not registered for this license


### GET `/productver/:product_guid`

Check for product updates.

**Response:**
```json
{
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "LatestVersion": "5.5.1.0",
  "DownloadURL": "https://example.com/downloads/product-5.5.1.zip"
}
```

---

# 2. Admin REST API

**Purpose:** Programmatic administration of customers, products, licenses, features, and machine registrations. Used 
by automation tools, scripts, or external systems.

## Authentication

Admin endpoints use the `X-API-Key` header:

```
GET /api/admin/customers HTTP/1.1
Host: localhost:8080
X-API-Key: your-admin-api-key
```

See [Authentication Configuration](#authentication-configuration) for setup details.

## Endpoints

### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/customers` | List all customers |
| GET | `/api/admin/customers/:id` | Get a customer |
| POST | `/api/admin/customers` | Create a customer |
| PUT | `/api/admin/customers/:id` | Update a customer |
| DELETE | `/api/admin/customers/:id` | Delete a customer |
| GET | `/api/admin/customers/:id/exists` | Check if customer exists |

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/products` | List all products |
| GET | `/api/admin/products/:id` | Get a product |
| POST | `/api/admin/products` | Create a product |
| PUT | `/api/admin/products/:id` | Update a product |
| DELETE | `/api/admin/products/:id` | Delete a product |

### Customer Products (Licenses)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/customers/:customerId/products` | List customer's licenses |
| GET | `/api/admin/customers/:customerId/unlicensed-products` | List products not yet licensed |
| POST | `/api/admin/customers/:customerId/products` | Create a license |
| PUT | `/api/admin/customers/:customerId/products/:productId` | Update a license |
| DELETE | `/api/admin/customers/:customerId/products/:productId` | Delete a license |

### Features (Product Feature Definitions)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/products/:productId/features` | List product's features |
| POST | `/api/admin/products/:productId/features` | Create a feature |
| PUT | `/api/admin/features/:id` | Update a feature |
| DELETE | `/api/admin/features/:id` | Delete a feature |

### Product Features (Customer-Specific Values)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/customers/:customerId/products/:productId/features` | Get customer's feature values |
| PUT | `/api/admin/customers/:customerId/products/:productId/features/:id` | Update a feature value |

**Storage Design:** Feature values use an override-only pattern. The `license_feature` table only stores values that 
differ from the feature's default. When no override exists, the default value from the `feature` table is used. This 
keeps the database lean and makes it easy to change defaults globally.

### Machine Registrations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/customers/:customerId/products/:productId/registrations` | List machine registrations |
| DELETE | `/api/admin/registrations/:machineId/:productId` | Delete a machine registration |

Query parameters for machine registrations:
- `active=true` - Only return active (non-expired) registrations

### Expirations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/expirations` | List expired licenses |

Query parameters:
- `before=yyyy-mm-dd` - Show licenses where either expiration date or maintenance expiration date is before this date (default: today)

**Response:**
```json
[
  {
    "customerName": "Acme Corp",
    "contactName": "John Doe",
    "email": "john@acme.com",
    "productName": "Widget Pro",
    "expirationDate": "2024-12-15",
    "maintExpirationDate": "2024-12-15"
  }
]
```

Results include licenses where either the license expiration date OR the maintenance expiration date is before the
specified date. Results are sorted by expiration date descending (most recently expired first).

### Backup

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/backup` | Create a database backup |

**Response:**
```json
{
  "filename": "2025-01-09_16.30.00_regdump.sql.gz",
  "path": "/path/to/db/backups/2025-01-09_16.30.00_regdump.sql.gz",
  "size": 4523
}
```

Creates a gzip-compressed SQL dump of the database. Backups are saved to a `backups/` directory relative to the database file. The SQL dump includes all schema definitions and data, wrapped in a transaction for safe restoration.

**To restore a backup:**
```bash
# Linux/macOS
gunzip -c 2025-01-09_16.30.00_regdump.sql.gz | sqlite3 restored.db

# Windows (PowerShell)
7z e -so backup.sql.gz | sqlite3 restored.db
```

---

# 3. Admin Web Frontend

**Purpose:** Browser-based administration interface for managing customers, products, licenses, and registrations. 
Provides the same functionality as the Admin API with a user-friendly UI.

## Tech Stack

- **Templ** - Type-safe Go HTML templates that compile to Go code
- **HTMX** - Dynamic interactions without JavaScript frameworks
- **DaisyUI/Tailwind CSS** - Loaded via CDN, no build step needed
- **Tom Select** - Searchable dropdowns for large lists (via CDN)

## Access

Navigate to `/web/` in your browser. You'll be redirected to a login page where you enter your `ADMIN_API_KEY`. 
See [Authentication Configuration](#authentication-configuration) for details.

## Features

- **Registrations** - Customer selector with registration overview
- **Customer Management** - Create, edit, delete customers
- **Product Catalog** - Manage products and their feature definitions
- **License Management** - Assign products to customers with seat counts, terms, and expiration dates
- **Feature Values** - Configure customer-specific feature values (integer, string, or enum types)
- **Machine Registrations** - View and manage individual machine activations
- **Offline Registration** - Manual registration for customers without internet access
- **Database Backup** - One-click backup from the sidebar (creates timestamped gzip-compressed SQL dump)

## Routes

| Route | Description |
|-------|-------------|
| `/web/login` | Login page |
| `/web/` | Licenses with customer selector |
| `/web/customers` | Customer list and management |
| `/web/products` | Product catalog and feature definitions |
| `/web/licenses/:customerID` | Customer's product licenses |
| `/web/features/:customerID/:productID` | Feature value configuration |
| `/web/machines/:customerID/:productID` | Machine registration list |
| `/web/backup` | Create database backup (POST) |

## Offline Registration

For customers without internet access, the admin can manually create registrations and export them as JSON files that can be 
loaded by the client software.

### Workflow

1. **Get Machine Code from Customer** - The customer runs the client software, which displays their unique machine code. 
   They communicate this to the admin (email, phone, etc.).

2. **Create Manual Registration** - In the admin web UI:
   - Navigate to the customer's registrations
   - Click the desktop icon to open Machine Registrations
   - Click **Add** to open the "Manual Registration (Offline)" form
   - Enter the machine code and user name
   - Click **OK** to create the registration

3. **Export Registration File** - Click the download icon next to the machine entry to export a JSON file containing the 
   complete registration data (same format as the `/api/v1/activate` response).

4. **Transfer to Customer** - Send the JSON file to the customer via email, USB drive, or other means.

5. **Load in Client Software** - The customer loads the registration file in their client software using its "load registration 
   from file" option.

### Exported JSON Format

The exported file matches the standard activation response:

```json
{
  "UserName": "Joe User",
  "UserCompany": "Acme Corp",
  "MachineCode": "5mToXAaMQRRXOG58VT2oRKBgD8c=...",
  "ExpirationDate": "2025-12-31",
  "MaintExpirationDate": "2025-12-31",
  "MaxProductVersion": "",
  "LatestVersion": "5.5.1.0",
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "LicenseKey": "287d3e24-af8e-4f45-99e8-a9e9f1aa1a91",
  "RegistrationHash": "ogwa5eQEaFjN/28bsapgee3cyH0=",
  "Features": {}
}
```

The `RegistrationHash` is computed identically to online activations, so the client software can validate it using the same logic.

## Building

The web UI requires the `templ` CLI to generate Go code from `.templ` template files:

```bash
# Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Generate templates
templ generate

# Build everything
task build
```

---

# Authentication Configuration

The server has three authentication mechanisms:

| API | Method | Purpose |
|-----|--------|---------|
| Client API (`/api/v1/*`) | `X-License-Key` header | Customer license key (validated against database) |
| Admin API (`/api/admin/*`) | `X-API-Key` header | Server admin key (validated against env var) |
| Web UI (`/web/*`) | Login page + cookie | Browser-based authentication |

## Admin API Key

The `ADMIN_API_KEY` environment variable is **required**. The server will refuse to start if it is not set.

### Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Set your API key in `.env`:
   ```
   ADMIN_API_KEY=your-secret-here
   ```

   This can be any string (it's compared like a password). For better security, generate a random value:
   ```bash
   task gen:apikey
   ```

3. Source the environment (or configure your IDE/shell):
   ```bash
   # Linux/macOS
   export $(cat .env | xargs)

   # PowerShell
   Get-Content .env | ForEach-Object { if ($_ -match '^([^#][^=]+)=(.*)$') { [Environment]::SetEnvironmentVariable($matches[1], $matches[2]) } }
   ```

### Admin API Requests

Admin API requests require the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-secure-api-key-here" http://localhost:8080/api/admin/customers
```

### Web UI Authentication

The web UI uses session-based authentication with CSRF protection:

1. Navigate to `/web/` - redirects to login if not authenticated
2. Enter your `ADMIN_API_KEY` value
3. Server creates an in-memory session and returns a session ID cookie (7 days)
4. All subsequent requests validated via session ID lookup
5. Click "Logout" in sidebar to end session

**Security features:**
- **Session-based**: Only the session ID is stored in the cookie, never the API key
- **CSRF protection**: All mutating requests require a CSRF token
- **HttpOnly cookie**: Session cookie not accessible via JavaScript
- **Secure cookie**: Only sent over HTTPS (when using TLS)
- **SameSite=Strict**: Prevents CSRF attacks from external sites

**Note:** Sessions are stored in memory and do not persist across server restarts. Users will need to log in again after a restart.

### Environment Variable Summary

| Variable | Required | Description |
|----------|----------|-------------|
| `ADMIN_API_KEY` | **Yes** | Admin API and Web UI authentication key |
| `REGISTRATION_SECRET` | **Yes** | Secret key appended before hashing registration data |
| `DB_PATH` | No | Database file path (overrides `db_path` in config.yaml) |
| `PORT` | No | Server port (overrides `addr` in config.yaml, useful for cloud platforms) |

**⚠️ Warning:** Changing `REGISTRATION_SECRET` after deployment will invalidate all existing registrations. Every client 
will need to re-activate their license. The same secret must also be used by client software when validating registration 
files offline.

### Database Configuration

The database path is determined in this order:
1. `DB_PATH` environment variable (if set)
2. `db_path` in `config.yaml` (default: `./testdata/registrations.db`)
3. Fallback: `./registrations.db`

---

# Deployment

## Local Development

```bash
# Run directly
./dist/regserver

# Or with Caddy for local HTTPS
task dev
```

## Database Creation

Database "migrations" are applied automatically when the server starts. If no database exists at the configuration settings, 
one will be created and schema applied.

Note that all GUID columns are created with COLLATE NOCASE (for case-insensitive comparisons) and converted to lower-case
when saved to the database.

**Cascade Deletes:** Deleting a customer removes all their machines, licenses, and registrations. Deleting a product
removes all licenses, features, and registrations for that product. This is enforced at the database level.

## Production (DigitalOcean / Linux VPS)

For production deployment with `Caddy` and `systemd`, see **[DEPLOYMENT.md](doc/DEPLOYMENT.md)** for a complete step-by-step
guide covering:

- Building Linux binaries
- Installing and configuring Caddy
- Setting up systemd services
- Environment variable configuration
- Firewall and security setup

## Automated Backups

The `scripts/backup-to-cloud.sh` script automates database backups with optional cloud upload using [rclone](https://rclone.org/).

### Summary of backup workflow:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Cron/Systemd   │────▶│  backup-to-     │────▶│  Backblaze B2   │
│  Timer          │     │  cloud.sh       │     │  (via rclone)   │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  POST /api/     │
                        │  admin/backup   │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Local .sql.gz  │
                        │  backup file    │
                        └─────────────────┘
```

### Prerequisites

```bash
# Install rclone (Linux)
curl https://rclone.org/install.sh | sudo bash

# Or via package manager
sudo apt install rclone    # Debian/Ubuntu
brew install rclone        # macOS
```

### Configure Cloud Storage (Backblaze B2)

**1. Create B2 bucket and application key:**

1. Log in to [Backblaze B2 Cloud Storage](https://secure.backblaze.com/b2_buckets.htm)
2. Create a new bucket (e.g., `mycompany-regserver-backups`)
   - Set to **Private**
3. Go to **App Keys** and click **Add a New Application Key**
   - Name: `regserver-backup`
   - Allow access to bucket: select your bucket
   - Type of access: **Read and Write**
4. **Save the keyID and applicationKey** - the application key is only shown once!

**2. Configure rclone with your B2 credentials:**

```bash
rclone config
```

Follow the prompts:
```
n) New remote
name> b2
Storage> b2  (or type the number for "Backblaze B2")
account> YOUR_KEY_ID        # e.g., 0012345678abcdef0000000001
key> YOUR_APPLICATION_KEY   # e.g., K001xxxxxxxxxxxxxxxxxxxxxxxxxxxx
hard_delete> (leave default)
Edit advanced config> n
```

**3. Test the connection:**

```bash
# List buckets
rclone lsd b2:

# List contents of your backup bucket
rclone ls b2:mycompany-regserver-backups
```

### Environment Setup

Add to your `.env` file:

```bash
# Required
ADMIN_API_KEY=your-admin-key

# For cloud upload
RCLONE_REMOTE=b2:your-bucket/regserver-backups

# Optional
REGSERVER_URL=http://localhost:8080   # Default
BACKUP_RETAIN_DAYS=7                   # Days to keep local backups
```

### Manual Usage

```bash
# Local backup only
./scripts/backup-to-cloud.sh

# Backup and upload to cloud
./scripts/backup-to-cloud.sh --upload

# Backup, upload, and prune old local files
./scripts/backup-to-cloud.sh --upload --prune
```

### Cron Job Setup

Schedule automated backups using cron:

```bash
# Edit crontab
crontab -e

# Add one of these lines:

# Daily at 2:00 AM - backup and upload
0 2 * * * /opt/regserver/scripts/backup-to-cloud.sh --upload >> /var/log/regserver-backup.log 2>&1

# Daily at 2:00 AM - backup, upload, and prune local files older than 7 days
0 2 * * * /opt/regserver/scripts/backup-to-cloud.sh --upload --prune >> /var/log/regserver-backup.log 2>&1

# Every 6 hours - more frequent backups
0 */6 * * * /opt/regserver/scripts/backup-to-cloud.sh --upload >> /var/log/regserver-backup.log 2>&1
```

**Important:** Ensure the cron environment has access to required variables. Either:

1. Source the .env file in your script (already handled by the script)
2. Or define variables in crontab:
   ```bash
   ADMIN_API_KEY=your-key
   RCLONE_REMOTE=b2:bucket/path
   0 2 * * * /opt/regserver/scripts/backup-to-cloud.sh --upload
   ```

### Systemd Timer Alternative

For systemd-based systems, you can use a timer instead of cron:

```bash
# /etc/systemd/system/regserver-backup.service
[Unit]
Description=RegServer Database Backup
After=network.target

[Service]
Type=oneshot
ExecStart=/opt/regserver/scripts/backup-to-cloud.sh --upload --prune
EnvironmentFile=/opt/regserver/.env
User=regserver

# /etc/systemd/system/regserver-backup.timer
[Unit]
Description=Daily RegServer Backup

[Timer]
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable the timer:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now regserver-backup.timer

# Check status
sudo systemctl list-timers | grep regserver
```

### Restore from Cloud Backup

```bash
# List available backups
rclone ls b2:your-bucket/regserver-backups

# Download a specific backup
rclone copy b2:your-bucket/regserver-backups/2025-01-09_02.00.00_regdump.sql.gz ./

# Restore to new database
gunzip -c 2025-01-09_02.00.00_regdump.sql.gz | sqlite3 restored.db
```

## Demo Mode

The `-demo` flag loads sample data when creating a new database. This is useful for:

- Trying out the software without manual data entry
- Running a public demo on platforms with ephemeral storage (like Render.com free tier)

```bash
./regserver -demo
```

**Sample data includes:**

| Customer | Product | License Type | Seats |
|----------|---------|--------------|-------|
| Acme Corporation | DataMapper Pro | Perpetual + maintenance | 25 |
| Acme Corporation | ReportBuilder | Perpetual | 10 |
| TechStart Inc | DataMapper Pro | Annual subscription | 10 |
| Global Industries | DataMapper Pro | Perpetual, version-locked | 50 |
| Global Industries | ReportBuilder | Monthly subscription | 5 |

Demo license keys follow the pattern `11111111-1111-1111-1111-111111111111` through `55555555-...`.

### Render.com Deployment for Demo

The repository includes a `render.yaml` for one-click deployment to Render.com:

1. Push this repo to GitHub/GitLab
2. Connect the repo in Render dashboard
3. Render auto-detects `render.yaml` and deploys

**Default demo credentials:**
- **Admin login**: `demo`

Since Render's free tier uses ephemeral storage, the database resets on each restart (after ~15 min of inactivity). 
The `-demo` flag ensures fresh sample data is loaded each time.

**Demo Registration key for API testing**

```
REGISTRATION_SECRET=demo-secret-do-not-use-in-production
```

---

# Architecture

```
internal/
├── customer/           # Customer domain (model, repo, service)
├── product/            # Product domain
├── feature/            # Feature definitions
├── featurevalue/       # Customer-specific feature values
├── license/            # Customer-product licenses
├── machine/            # Machine tracking
├── registration/       # Machine-product registrations
├── activation/         # License activation logic
├── http/
│   ├── admin/          # Admin REST API handlers
│   ├── client/         # Client registration API handlers
│   └── web/            # Web UI handlers
├── middleware/         # Auth, sessions, CSRF, theme
├── server/             # Server builder
├── sqlite/             # Database migrations
└── viewmodels/         # View models for templates
```

---

# General Notes

- Registration hashes are generated using UTF-16LE encoding, SHA1 (with appended `REGISTRATION_SECRET`), and Base64 output for compatibility with legacy Delphi clients
- Activation handles both insert and update via `ON CONFLICT DO UPDATE`
- Expiration dates are sourced from `license.expiration_date` at activation time

---

# Developer Notes

## API Testing with Bruno

The project includes a [Bruno](https://www.usebruno.com/) collection for testing API endpoints. Bruno is a fast, free, 
git-friendly API client.

### Setup

```bash
# Install via Scoop (Windows)
scoop install bruno

# Or download from https://www.usebruno.com/downloads
```

### Usage

1. Open Bruno
2. Click "Open Collection" and select the `scripts/bruno/` folder
3. Select the "local" environment from the dropdown (top-right)
4. Update variables in `scripts/bruno/environments/local.bru` as needed:
   - `baseUrl` - Server URL (default: `http://localhost:8080`)
   - `licenseKey` - A valid `license_key` from the `license` table (via `SAMPLE_LICENSE_KEY` env var)

### Collection Structure

```
bruno/
├── bruno.json              # Collection manifest
├── environments/
│   └── local.bru           # Environment variables
├── admin/                  # Admin API requests (X-API-Key)
│   ├── list-customers.bru
│   └── list-products.bru
└── client/                 # Client API requests (X-License-Key)
    ├── activate.bru
    └── product-version.bru
```

The `admin/` and `client/` folders are separated to reflect the different authentication mechanisms:
- **Admin API** - Uses `X-API-Key` header (server admin)
- **Client API** - Uses `X-License-Key` header (customer software)

### Adding New Requests

Create a new `.bru` file in the appropriate folder (`admin/` or `client/`):

```
meta {
  name: Request Name
  type: http
  seq: 2
}

get {
  url: {{baseUrl}}/api/admin/endpoint
  body: none
  auth: none
}

headers {
  X-API-Key: {{adminKey}}
}
```

---

## Client Implementation

The Registration service (`/api/v1/activate`) is usually called only after checking a local registration
file and finding an invalid license (i.e. is missing or expired). 

The local registration file contains a hash of the other registration fields. This hash should be 
calculated each time the program runs and checked against the one saved in the local registration file.

Documentation and sample client implementations can be found in the [doc/clients](doc/clients) folder. 

---

## HTMX

HTMX is used by the `/web` admin pages for dynamic interactions without writing JavaScript.

### Main Patterns

| Pattern             | Usage                                                                           |
|---------------------|---------------------------------------------------------------------------------|
| Modal forms         | hx-get loads form into #modal-content, JS opens modal on htmx:afterSwap         |
| Table refresh       | Forms submit with hx-post/put/delete, target #*-table-container, swap innerHTML |
| Delete confirmation | hx-confirm="Are you sure?" shows browser confirm dialog                         |
| Partial responses   | Handlers return HTML fragments, not full pages                                  |

**Example: Edit Customer Button**

```html
<button
   hx-get="/web/customers/5/edit"       <!-- Fetch edit form -->
   hx-target="#modal-content"           <!-- Put response here -->
   hx-swap="innerHTML"                  <!-- Replace content -->
>Edit</button>
```

**Example: Create/Update Form**

```html
  <form
      hx-post="/web/customers"             <!-- POST for new -->
      hx-put="/web/customers/5"            <!-- PUT for update -->
      hx-target="#customers-table-container"
      hx-swap="innerHTML"
  >
```

**Flow**

1. Button click → HTMX fetches HTML fragment from server
2. Fragment injected into target element
3. For modals: htmx:afterSwap event triggers showModal() in app.js
4. Form submit → Server returns updated table HTML → Table refreshes in place

**Key Files**

- base.templ - Enables hx-ext="response-targets"
- app.js - Opens modal after HTMX swaps content into #modal-content
- Component templates - All HTMX attributes on buttons/forms

The `isHTMX()` helper checks for the HX-Request header so the handler can determine
if a full page or a fragment should be rendered. For example:

```go
 if isHTMX(c) {
      return components.CustomersTable(viewCustomers).Render(ctx, c.Response())
  }
  return pages.Customers(viewCustomers).Render(ctx, c.Response())
```
**Behavior:**

| Request Type              | Returns                                   |
|---------------------------|-------------------------------------------|
| Direct browser navigation | Full page (pages.Customers)               |
| HTMX request              | Fragment only (components.CustomersTable) |

This allows the same /web/customers endpoint to work for:
1. Bookmarking / page refresh → full page with layout
2. HTMX table refresh after CRUD → just the table component


### Modal Forms

1. Modal Container (in base.templ)

   ```html
     <dialog id="modal" class="modal">
         <div class="modal-box w-11/12 max-w-6xl max-h-[80vh] overflow-y-auto">
             <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2"
                     onclick="handleModalClose()">✕</button>
             <div id="modal-content">
                 <!-- Content loaded via HTMX -->
             </div>
         </div>
         <div class="modal-backdrop" onclick="handleModalClose()"></div>
     </dialog>
   ```

2. Button to Open Modal

   ```html
   <button
      hx-get="/web/customers/5/edit"
      hx-target="#modal-content"
      hx-swap="innerHTML"
   >Edit</button>
   ```

3. Form Template (returned by handler)

   ```html
      templ CustomerForm(customer *vm.Customer) {
      <h3 class="font-bold text-lg mb-4">Edit Customer</h3>
       <form
           hx-put="/web/customers/5"
           hx-target="#customers-table-container"
           hx-swap="innerHTML"
       >
           <!-- fields -->
           <div class="modal-action">
               <button type="button" class="btn" onclick="closeModal()">Cancel</button>
               <button type="submit" class="btn btn-primary">Save</button>
           </div>
       </form>
   }
   ```

4. JavaScript (app.js)

   ```javascript
   // Auto-open modal when content is loaded
   document.body.addEventListener('htmx:afterSwap', function(evt) {
     if (evt.detail.target.id === 'modal-content') {
       showModal();
       // Handle nested modals with back navigation
       var backUrlEl = document.querySelector('#modal-content [data-init-back-url]');
       if (backUrlEl) {
         document.getElementById('modal-content').setAttribute('data-back-url', backUrlEl.getAttribute('data-back-url'));
         backUrlEl.remove();
       } else {
         document.getElementById('modal-content').removeAttribute('data-back-url');
       }
     }
   });

   // Close modal on success (triggered by HX-Trigger header)
   document.body.addEventListener('closeModal', function() {
     closeModal();
   });

   // Show toast notifications (triggered by HX-Trigger header)
   document.body.addEventListener('showToast', function(evt) {
     showToast(evt.detail.message, evt.detail.type);
   });
   ```

5. Handler Response (on success)

   ```go
   setTriggerWithData(c, `{"closeModal": true, "showToast": {"message": "Saved!", "type": "success"}}`)
   return components.CustomersTable(customers).Render(ctx, c.Response())
   ```

**Flow:**

1. Click Edit → HTMX fetches form into #modal-content
2. htmx:afterSwap fires → JS calls showModal()
3. Submit → HTMX posts form, targets the table container
4. Handler returns updated table + HX-Trigger header
5. HTMX swaps table, fires closeModal event → JS closes modal

---

# License

This project is licensed under the **O'Saasy License**. Basically… the MIT do-whatever-you-want license, 
but with the commercial rights for SaaS reserved for the copyright holder.

---

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
