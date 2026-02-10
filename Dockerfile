# Build stage
FROM golang:1.26.0-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
# CGO_ENABLED=0 for static binary compatible with scratch
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bubble .

# Runtime stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/bubble /bubble

# Copy CA certificates for HTTPS (if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the default port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/bubble"]
