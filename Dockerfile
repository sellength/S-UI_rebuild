# ==========================================
# Phase 1: Build Frontend
# ==========================================
FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install --legacy-peer-deps
COPY frontend/ ./
RUN npm run build

# ==========================================
# Phase 2: Build Go Backend
# ==========================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download
WORKDIR /app
COPY backend ./backend
COPY --from=frontend-builder /app/dist ./backend/web/html
WORKDIR /app/backend
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags postgres -ldflags "-s -w" -o /out/sui main.go

# ==========================================
# Phase 2.5: Build Sing-Box with tags
# ==========================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS singbox-builder
ARG TARGETOS
ARG TARGETARCH
ARG SINGBOX_VER=1.13.13
RUN apk add --no-cache git
RUN git clone --branch v${SINGBOX_VER} --depth 1 https://github.com/sagernet/sing-box.git /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags with_v2ray_api -ldflags "-s -w" -o /out/sing-box ./cmd/sing-box

# ==========================================
# Phase 3: Final Runtime Image
# ==========================================
FROM alpine:3.22
LABEL org.opencontainers.image.authors="sellength" \
      org.opencontainers.image.description="s-ui: Advanced Dual-Stack Sing-box User Interface Dashboard"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    openssl \
    curl \
    bash \
    git

# Set system default timezone
ENV TZ=Asia/Shanghai
ENV SUI_ACME_SH=/usr/local/s-ui/acme/acme.sh
ENV SUI_ACME_HOME=/usr/local/s-ui/acme
ENV SUI_CERTIFICATE_WORK_DIR=/usr/local/s-ui/certificates

WORKDIR /usr/local/s-ui

# Copy compiled sui binary
COPY --from=backend-builder /out/sui /usr/local/s-ui/sui

# Clone acme.sh
RUN git clone --depth 1 https://github.com/acmesh-official/acme.sh.git /usr/local/s-ui/acme \
    && chmod +x /usr/local/s-ui/acme/acme.sh

# Ensure runtime directory structures exist
RUN mkdir -p \
    /usr/local/s-ui/db \
    /usr/local/s-ui/bin \
    /usr/local/s-ui/bin_tmpl \
    /usr/local/s-ui/cert \
    /usr/local/s-ui/certificates

# Copy runSingbox.sh to templates directory
COPY core/runSingbox.sh /usr/local/s-ui/bin_tmpl/runSingbox.sh

# Copy compiled static sing-box binary matching target architecture
COPY --from=singbox-builder /out/sing-box /usr/local/s-ui/bin_tmpl/sing-box

RUN chmod +x /usr/local/s-ui/sui /usr/local/s-ui/bin_tmpl/runSingbox.sh /usr/local/s-ui/bin_tmpl/sing-box

# Expose s-ui standard dashboard and subscription ports
EXPOSE 2095 2096

# CMD copies binaries from template folder to named volume, runs migrations, and starts sui
CMD ["/bin/sh", "-c", "cp -f /usr/local/s-ui/bin_tmpl/* /usr/local/s-ui/bin/ && /usr/local/s-ui/sui migrate && /usr/local/s-ui/sui"]
