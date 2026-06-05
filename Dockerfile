# ========================================================
# Phase 1: Compile s-ui binary in Go environment
# ========================================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /build

# Copy backend source codes including web static assets
COPY backend/ .

# Ensure dependencies are downloaded
RUN go mod download

# Target architecture parameters passed by Docker buildx
ARG TARGETOS
ARG TARGETARCH

# Compile static linked binary for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o sui main.go

# ========================================================
# Phase 2: Minimal and Secure Runtime Environment for s-ui
# ========================================================
FROM alpine:latest

LABEL org.opencontainers.image.authors="sellength" \
      org.opencontainers.image.description="s-ui: Advanced Dual-Stack Sing-box User Interface Dashboard (Rebuild)"

# Install essential runtime dependencies using lightweight apk
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    curl \
    bash

# Set system default timezone
ENV TZ=Asia/Shanghai

WORKDIR /usr/local/s-ui

# Copy s-ui binary from builder phase
COPY --from=builder /build/sui /usr/local/s-ui/sui

# Ensure database, cert and bin directory structures exist with proper permissions
RUN mkdir -p /usr/local/s-ui/db /usr/local/s-ui/bin /usr/local/s-ui/cert && \
    chmod +x /usr/local/s-ui/sui

# Expose s-ui standard dashboard administration port
EXPOSE 2095

# Command to execute migrations first, then run panel main process
CMD ["/bin/sh", "-c", "/usr/local/s-ui/sui migrate && /usr/local/s-ui/sui"]