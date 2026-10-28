#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

# Run from the repository root on the server. Only an age public recipient is
# needed here; the decryption identity belongs outside the VPS and repository.
: "${AGE_RECIPIENT:?Set the public age recipient before backing up}"
: "${BACKUP_DIRECTORY:?Set an absolute private backup directory}"
case "$BACKUP_DIRECTORY" in /*) ;; *) echo 'Backup directory must be absolute' >&2; exit 2;; esac
for executable in docker age mktemp sha256sum; do command -v "$executable" >/dev/null; done
mkdir -p -- "$BACKUP_DIRECTORY"
temporary=$(mktemp "$BACKUP_DIRECTORY/.andaria-encrypted.XXXXXXXX")
trap 'rm -f -- "$temporary"' EXIT
compose=(docker compose --env-file .env.production -p diplomado-andaria -f docker-compose.prod.yml)
"${compose[@]}" exec -T postgres sh -c 'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom --no-owner --no-acl' \
  | age --encrypt --recipient "$AGE_RECIPIENT" > "$temporary"
test -s "$temporary"
destination="$BACKUP_DIRECTORY/andaria-$(date -u +%Y%m%dT%H%M%SZ)-$(basename "$temporary").dump.age"
mv -- "$temporary" "$destination"
sha256sum "$destination" > "$destination.sha256"
echo 'Encrypted backup and checksum created. Copy them to the approved off-server destination.'
