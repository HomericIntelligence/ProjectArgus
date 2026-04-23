#!/usr/bin/env bash
set -euo pipefail

# Back up Prometheus and Loki data volumes to timestamped tar.gz files
# Usage: backup.sh

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="$(cd "$(dirname "$0")/.." && pwd)/backups"
mkdir -p "$BACKUP_DIR"

echo "Backing up Prometheus and Loki data volumes..."
echo "Backup directory: $BACKUP_DIR"

for vol in prometheus_data loki_data; do
    out="$BACKUP_DIR/${vol}_${TIMESTAMP}.tar.gz"
    echo "Backing up $vol -> $out"
    docker run --rm -v "${vol}:/data:ro" -v "${BACKUP_DIR}:/backups" \
        alpine tar czf "/backups/${vol}_${TIMESTAMP}.tar.gz" -C /data .
    echo "  ✓ $vol backup complete"
done

echo ""
echo "Backup complete: $BACKUP_DIR"
echo "Files:"
ls -lh "$BACKUP_DIR/${TIMESTAMP:0:8}"* 2>/dev/null || echo "  (no backups found with today's date)"
