#!/bin/bash
#
# backup-to-cloud.sh - Create database backup and sync to cloud storage
#
# This script calls the regserver backup API and optionally syncs the backup
# directory to a cloud storage provider using rclone.
#
# Prerequisites:
#   - curl
#   - rclone (if uploading to cloud) - https://rclone.org/install/
#
# Usage:
#   ./backup-to-cloud.sh                    # Local backup only
#   ./backup-to-cloud.sh --upload           # Backup and upload to cloud
#   ./backup-to-cloud.sh --upload --prune   # Backup, upload, and prune old local files
#
# Environment variables (can also be set in .env file):
#   REGSERVER_URL      - Server URL (default: http://localhost:8080)
#   ADMIN_API_KEY      - Required for API authentication
#   BACKUP_DIR         - Local backup directory (default: auto-detected from API response)
#   RCLONE_REMOTE      - rclone remote and path (e.g., "b2:mybucket/backups")
#   BACKUP_RETAIN_DAYS - Days to keep local backups when pruning (default: 7)
#

set -e

# Load .env file if it exists
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/../.env" ]; then
    export $(grep -v '^#' "$SCRIPT_DIR/../.env" | xargs)
fi

# Configuration with defaults
REGSERVER_URL="${REGSERVER_URL:-http://localhost:8080}"
BACKUP_RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"

# Parse arguments
UPLOAD=false
PRUNE=false
for arg in "$@"; do
    case $arg in
        --upload)
            UPLOAD=true
            ;;
        --prune)
            PRUNE=true
            ;;
        --help|-h)
            head -30 "$0" | tail -25
            exit 0
            ;;
    esac
done

# Validate required settings
if [ -z "$ADMIN_API_KEY" ]; then
    echo "Error: ADMIN_API_KEY environment variable is required"
    exit 1
fi

if [ "$UPLOAD" = true ] && [ -z "$RCLONE_REMOTE" ]; then
    echo "Error: RCLONE_REMOTE environment variable is required for --upload"
    echo "Example: RCLONE_REMOTE=b2:mybucket/regserver-backups"
    exit 1
fi

if [ "$UPLOAD" = true ] && ! command -v rclone &> /dev/null; then
    echo "Error: rclone is not installed. Install from https://rclone.org/install/"
    exit 1
fi

echo "$(date '+%Y-%m-%d %H:%M:%S') - Starting backup..."

# Call backup API
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "X-API-Key: $ADMIN_API_KEY" \
    "$REGSERVER_URL/api/admin/backup")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "200" ]; then
    echo "Error: Backup API returned HTTP $HTTP_CODE"
    echo "$BODY"
    exit 1
fi

# Parse response
FILENAME=$(echo "$BODY" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)
FILEPATH=$(echo "$BODY" | grep -o '"path":"[^"]*"' | cut -d'"' -f4)
SIZE=$(echo "$BODY" | grep -o '"size":[0-9]*' | cut -d':' -f2)

echo "Backup created: $FILENAME ($SIZE bytes)"

# Upload to cloud if requested
if [ "$UPLOAD" = true ]; then
    BACKUP_DIR=$(dirname "$FILEPATH")
    echo "Syncing to $RCLONE_REMOTE..."
    rclone copy "$BACKUP_DIR" "$RCLONE_REMOTE" --progress
    echo "Upload complete"
fi

# Prune old local backups if requested
if [ "$PRUNE" = true ]; then
    BACKUP_DIR=$(dirname "$FILEPATH")
    echo "Pruning local backups older than $BACKUP_RETAIN_DAYS days..."
    find "$BACKUP_DIR" -name "*.sql.gz" -type f -mtime +$BACKUP_RETAIN_DAYS -delete -print
    echo "Prune complete"
fi

echo "$(date '+%Y-%m-%d %H:%M:%S') - Backup finished successfully"
