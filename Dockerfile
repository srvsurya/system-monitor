# ─────────────────────────────────────────────
# Stage 1: Frontend build (Node)
# ─────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /app

RUN mkdir -p embedfs/dist

WORKDIR /app/frontend

# Cache npm install layer separately
COPY frontend/package*.json ./
RUN npm ci

# Copy source and build into embedfs/dist
COPY frontend/ ./
RUN npm run build

# ─────────────────────────────────────────────
# Stage 2: Go build
# ─────────────────────────────────────────────
FROM golang:1.26-alpine AS go-builder

# gcc/musl needed for modernc.org/sqlite (CGO-free pure Go, but needs build tools)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy everything
COPY . .

# Pull in the built frontend from Stage 1
COPY --from=frontend-builder /app/embedfs/dist ./embedfs/dist

# Build the binary — release mode, statically linked
ENV GIN_MODE=release
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o system-monitor ./cmd/server

# ─────────────────────────────────────────────
# Stage 3: Minimal runtime image
# ─────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates for any outbound HTTPS (e.g. future webhooks)
RUN apk add --no-cache ca-certificates tzdata

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the binary
COPY --from=go-builder /app/system-monitor .

# SQLite DB lives here — will be a named volume
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

# Port your Gin server listens on
EXPOSE 8080

# Healthcheck — hits your existing /api/health or /api/metrics

ENTRYPOINT ["./system-monitor"]