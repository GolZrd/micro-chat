#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE auth_db OWNER "$POSTGRES_USER";
    CREATE DATABASE chat_db OWNER "$POSTGRES_USER";
EOSQL

echo "✅ Databases auth_db and chat_db created successfully!"