# Build stage
FROM golang:1.20-alpine AS builder

LABEL version="0.1.0-go"

WORKDIR /app

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy only the necessary source code (exclude development tools)
COPY cmd/tempdocs/ ./cmd/tempdocs/
COPY pkg/ ./pkg/

# Build the application with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/tempdocs cmd/tempdocs/main.go

# Final stage - use distroless for minimal image size and security
FROM gcr.io/distroless/static:nonroot

# Set environment variables
ENV SOURCE_DIR=/app/templates \
    OUTPUT_DIR=/app/docs/output \
    FORMAT=html \
    VERBOSE=false \
    VALIDATE_ONLY=false

# Create app directories
WORKDIR /app
COPY --from=builder /app/tempdocs /app/

# Documentation on expected volumes
VOLUME ["/app/templates", "/app/docs"]

# Expose port (if adding a server component in the future)
EXPOSE 8000

# Define entrypoint that accepts args
ENTRYPOINT ["/app/tempdocs"]
CMD ["--source", "/app/templates", "--output", "/app/docs/output", "--format", "html"] 