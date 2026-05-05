#!/usr/bin/env bash
set -euo pipefail

# Restore a Docker volume from a backup file
# Usage: restore.sh <volume_name> <backup_file.tar.gz>
# Example: restore.sh prometheus_data ./backups/prometheus_data_20260422_193000.tar.gz

if [[ $# -lt 2 ]]; then
    echo "Usage: restore.sh <volume_name> <backup_file.tar.gz>"
    echo ""
    echo "Example:"
    echo "  restore.sh prometheus_data ./backups/prometheus_data_20260422_193000.tar.gz"
    echo "  restore.sh loki_data ./backups/loki_data_20260422_193000.tar.gz"
    exit 1
fi

VOLUME="$1"
FILE="$2"

if [[ ! -f "$FILE" ]]; then
    echo "Error: Backup file not found: $FILE"
    exit 1
fi

echo "Restoring $VOLUME from $FILE ..."
${CONTAINER_CMD:-docker} run --rm -v "${VOLUME}:/data" -v "$(realpath "$FILE"):/backup.tar.gz:ro" \
    alpine sh -c "cd /data && tar xzf /backup.tar.gz"
echo "✓ Restore complete: $VOLUME"
