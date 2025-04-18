FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY cmd/tempdocs/ ./cmd/tempdocs/
COPY pkg/ ./pkg/
COPY internal/ ./internal/

ARG VERSION=1.0.0
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/ka1ne/template-doc-gen/internal/version.Version=${VERSION} \
    -X github.com/ka1ne/template-doc-gen/internal/version.Commit=${COMMIT} \
    -X github.com/ka1ne/template-doc-gen/internal/version.BuildDate=${BUILD_DATE} \
    -X github.com/ka1ne/template-doc-gen/internal/version.GoVersion=$(go version | cut -d ' ' -f 3)" \
    -o /app/tempdocs cmd/tempdocs/main.go

FROM gcr.io/distroless/static:nonroot

ENV SOURCE_DIR=/app/templates \
    OUTPUT_DIR=/app/docs/output \
    FORMAT=html \
    VERBOSE=false \
    VALIDATE_ONLY=false

WORKDIR /app
COPY --from=builder /app/tempdocs /app/

VOLUME ["/app/templates", "/app/docs"]
EXPOSE 8000

ENTRYPOINT ["/app/tempdocs"]
CMD ["--source", "/app/templates", "--output", "/app/docs/output", "--format", "html"] 