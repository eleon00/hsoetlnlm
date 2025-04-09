# Stage 1: Build the Go application and install rpk (Redpanda Connect)
FROM golang:1.24.2-alpine AS builder

# Install git, curl, unzip, needed for go mod download and rpk install
RUN apk add --no-cache git curl unzip ca-certificates

WORKDIR /app

# Copy module files and download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /hsoetlnlm cmd/server/main.go

# Download and install rpk (which includes rpk connect)
# See: https://github.com/redpanda-data/connect
# Pinning a version might be safer in the long run
RUN curl -LO https://github.com/redpanda-data/redpanda/releases/latest/download/rpk-linux-amd64.zip && \
    unzip rpk-linux-amd64.zip -d /usr/local/bin && \
    rm rpk-linux-amd64.zip && \
    chmod +x /usr/local/bin/rpk

# Stage 2: Create the final, minimal image
FROM alpine:latest

# Install ca-certificates for TLS connections (e.g., to Temporal Cloud, DBs)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the built application binary
COPY --from=builder /hsoetlnlm /app/hsoetlnlm

# Copy the rpk binary (includes Redpanda Connect)
COPY --from=builder /usr/local/bin/rpk /usr/local/bin/rpk

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["/app/hsoetlnlm"] 