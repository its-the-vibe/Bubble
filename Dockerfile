# Build stage
FROM --platform=$BUILDPLATFORM golang:1.27.0-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
# CGO_ENABLED=0 for static binary compatible with scratch
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -ldflags="-w -s" -o bubble .

# Runtime stage
FROM gcr.io/distroless/static-debian13:nonroot

# Copy the binary from builder
COPY --from=builder /build/bubble /bubble

# Expose the default port
EXPOSE 8080

USER nonroot:nonroot

# Run the application
ENTRYPOINT ["/bubble"]
