# Tahap build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ✅ Build untuk Linux Alpine
RUN go build -o main .

# ---

FROM alpine:latest

# ✅ Install tools
RUN apk add --no-cache nginx supervisor bash

WORKDIR /app

# ✅ Copy hasil build dan assets
COPY --from=builder /app/main .
RUN chmod +x /app/main
COPY --from=builder /app/dist/ticketing ./dist/ticketing
COPY --from=builder /app/nginx.conf /etc/nginx/nginx.conf

# ✅ Copy script wait-for-it
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# ✅ Buat direktori log
RUN mkdir -p /var/log/nginx /var/lib/nginx/tmp /run/nginx /var/log/supervisor

# ✅ Konfigurasi supervisor
RUN echo '[supervisord]' > /etc/supervisord.conf && \
    echo 'nodaemon=true' >> /etc/supervisord.conf && \
    echo '[program:nginx]' >> /etc/supervisord.conf && \
    echo 'command=/usr/sbin/nginx -g "daemon off;"' >> /etc/supervisord.conf && \
    echo 'autostart=true' >> /etc/supervisord.conf && \
    echo 'autorestart=true' >> /etc/supervisord.conf && \
    echo '[program:goapp]' >> /etc/supervisord.conf && \
    echo 'command=sh -c "/wait-for-it.sh super_apps:3306 -- /app/main"' >> /etc/supervisord.conf && \
    echo 'directory=/app' >> /etc/supervisord.conf && \
    echo 'autostart=true' >> /etc/supervisord.conf && \
    echo 'autorestart=true' >> /etc/supervisord.conf

# ✅ Expose port
EXPOSE 80

# ✅ Jalankan supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]