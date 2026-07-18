#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE identity_db;
    CREATE DATABASE transactions_db;
    CREATE DATABASE financing_db;
    CREATE DATABASE feedback_db;
    CREATE DATABASE ledger_db;
    CREATE DATABASE properties_db;
    CREATE DATABASE tokenization_db;
EOSQL
