# Tahap build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

# Tahap runtime (minimal)
FROM alpine:latest

WORKDIR /app

# Salin binary dari tahap builder
COPY --from=builder /app/main .

# Jalankan aplikasi
CMD ["./main"]
