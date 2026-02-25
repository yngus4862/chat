# filename: .devcontainer/postgres/initdb/00-create-keycloak.sh
#!/usr/bin/env sh
set -eu

echo "[initdb] create keycloak db/user (if not exists)"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${KEYCLOAK_DB_USER}') THEN
    CREATE ROLE ${KEYCLOAK_DB_USER} WITH LOGIN PASSWORD '${KEYCLOAK_DB_PASSWORD}';
  END IF;

  IF NOT EXISTS (SELECT FROM pg_database WHERE datname = '${KEYCLOAK_DB}') THEN
    CREATE DATABASE ${KEYCLOAK_DB} OWNER ${KEYCLOAK_DB_USER};
  END IF;
END
\$\$;
EOSQL
