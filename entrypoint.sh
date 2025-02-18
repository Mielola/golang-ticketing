#!/bin/bash

# Tunggu MySQL tersedia
until mysqladmin ping -h"$DB_HOST" --silent; do
  echo "Waiting for database connection..."
  sleep 2
done

# Jalankan migrasi database
echo "Running database migrations..."
/app/migrate -path ./db/migrations -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" up

# Jalankan aplikasi
echo "Starting backend service..."
exec "$@"
