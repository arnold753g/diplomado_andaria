#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

# Destructive operations are restricted to a newly created disposable database
# on a loopback PostgreSQL server. Never restores over the production database.
: "${RESTORE_CHECK_USER:?Set the local test PostgreSQL username}"
: "${AGE_IDENTITY_FILE:?Set the private age identity file outside the repository}"
: "${1:?Pass the encrypted backup file}"
for executable in age createdb dropdb pg_restore psql; do command -v "$executable" >/dev/null; done
database="andaria_restore_check_$(date -u +%Y%m%d%H%M%S)_${RANDOM}"
created=false
cleanup() {
  if "$created" && [[ "$database" =~ ^andaria_restore_check_[0-9]+_[0-9]+$ ]]; then
    dropdb --host=127.0.0.1 --username="$RESTORE_CHECK_USER" "$database"
  fi
}
trap cleanup EXIT
createdb --host=127.0.0.1 --username="$RESTORE_CHECK_USER" "$database"
created=true
age --decrypt --identity "$AGE_IDENTITY_FILE" "$1" \
  | pg_restore --host=127.0.0.1 --username="$RESTORE_CHECK_USER" --dbname="$database" --exit-on-error --no-owner --no-acl
psql --host=127.0.0.1 --username="$RESTORE_CHECK_USER" --dbname="$database" --set=ON_ERROR_STOP=1 --tuples-only \
  --command='SELECT max(version) FROM migraciones_esquema; SELECT count(*) FROM usuarios;' >/dev/null
echo 'Backup decrypted and restored into an isolated local database; basic schema checks passed.'
