FROM golang:1.24

RUN apt-get update && apt-get install -y default-mysql-client curl

# Mengunduh alat migration jika diperlukan
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar -xz -C /usr/local/bin

# Set working directory di dalam container
WORKDIR /app

# Menyalin file go.mod dan go.sum terlebih dahulu untuk mengunduh dependensi
COPY go.mod go.sum ./

# Download semua dependency
RUN go mod download

# Menyalin seluruh proyek ke dalam container
COPY . .

# Build aplikasi Go
RUN go build -o main main.go

# Memberikan izin eksekusi pada entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Expose port untuk aplikasi
EXPOSE 8080

# Menjalankan entrypoint dan aplikasi
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/main"]
