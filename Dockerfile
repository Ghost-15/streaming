# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies before copying source (faster rebuilds)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 → fully static binary compatible with distroless
# -ldflags='-w -s' → strip debug symbols (smaller binary, < 20 MB)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /streampulse \
    ./cmd/server

# ── Final stage — distroless ─────────────────────────────────────────────────
# gcr.io/distroless/static:nonroot → no shell, no package manager, minimal CVE surface
FROM gcr.io/distroless/static:nonroot

# Copy the binary from the builder stage
COPY --from=builder /streampulse /streampulse

EXPOSE 8080

ENTRYPOINT ["/streampulse"]
