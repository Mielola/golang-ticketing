#!/bin/bash

# Menunggu database siap (opsional, tergantung setup)
until mysql -h db -u root -p"$MYSQL_PASSWORD" -e "show databases" > /dev/null 2>&1; do
    echo "Waiting for database to be ready..."
    sleep 2
done

# Menjalankan migrasi jika ada
if [ -f "/app/migrate" ]; then
    echo "Running database migrations..."
    /usr/local/bin/migrate -path /app/migrations -database "mysql://root:$MYSQL_PASSWORD@tcp(db:3306)/commandcenter" up
else
    echo "Migrate tool not found, skipping migrations."
fi

# Menjalankan aplikasi utama
echo "Starting application..."
/app/main
