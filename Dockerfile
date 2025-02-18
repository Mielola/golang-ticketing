# Menggunakan base image Go
FROM golang:1.24

# Set working directory di dalam container
RUN mkdir /app
ADD . /app
WORKDIR /app

# Copy file go.mod dan go.sum terlebih dahulu
COPY go.mod go.sum ./

# Download semua dependency
RUN go mod download

# Copy seluruh isi direktori ke dalam container
COPY . .

# Build binary dari file main.go di root
RUN go build -o main main.go

# Expose port yang akan digunakan oleh aplikasi
EXPOSE 8080

# Perintah default untuk menjalankan aplikasi
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/main"]
