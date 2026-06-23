# ── Stage 1: Build ──────────────────────────────────────────
# Use Go image to compile — this image is heavy (~600MB)
# but never ships to production
FROM golang:1.25-alpine AS builder

WORKDIR /app

# copy dependency files first — Docker caches this layer
# if go.mod/go.sum unchanged, skips re-downloading on next build
COPY go.mod go.sum ./
RUN go mod download

# copy source code — this layer rebuilds only when code changes
COPY . .

# compile to a single binary — CGO disabled for pure Go binary
RUN CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

# ── Stage 2: Run ────────────────────────────────────────────
# Tiny alpine image (~5MB) — only the binary, no Go toolchain
FROM alpine:3.19

WORKDIR /app

# copy only the compiled binary from builder stage
COPY --from=builder /app/bin/server .

EXPOSE 8080
CMD ["./server"]