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
COPY --from=frontend-builder /app/frontend/dist ./backend/web/html
WORKDIR /app/backend
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags postgres -ldflags "-s -w" -o /out/sui main.go

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

# Download official sing-box binary matching target architecture
ARG TARGETARCH
ARG SINGBOX_VER=1.13.13
RUN set -ex && \
    ARCH="${TARGETARCH}" && \
    if [ -z "${ARCH}" ]; then \
        case "$(uname -m)" in \
            x86_64) ARCH="amd64" ;; \
            aarch64) ARCH="arm64" ;; \
            armv7*) ARCH="arm" ;; \
            i386|i686) ARCH="386" ;; \
        esac \
    fi && \
    ARCH_NAME="" && \
    case "${ARCH}" in \
        amd64) ARCH_NAME="amd64" ;; \
        arm64) ARCH_NAME="arm64" ;; \
        arm) ARCH_NAME="armv7" ;; \
        386) ARCH_NAME="386" ;; \
        *) echo "Unsupported arch: ${ARCH}" && exit 1 ;; \
    esac && \
    curl -Lo /tmp/sing-box.tar.gz "https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VER}/sing-box-${SINGBOX_VER}-linux-${ARCH_NAME}.tar.gz" && \
    tar -xzf /tmp/sing-box.tar.gz -C /tmp && \
    mv /tmp/sing-box-${SINGBOX_VER}-linux-${ARCH_NAME}/sing-box /usr/local/s-ui/bin_tmpl/sing-box && \
    rm -rf /tmp/sing-box*

RUN chmod +x /usr/local/s-ui/sui /usr/local/s-ui/bin_tmpl/runSingbox.sh /usr/local/s-ui/bin_tmpl/sing-box

# Expose s-ui standard dashboard and subscription ports
EXPOSE 2095 2096

# CMD copies binaries from template folder to named volume, runs migrations, and starts sui
CMD ["/bin/sh", "-c", "cp -f /usr/local/s-ui/bin_tmpl/* /usr/local/s-ui/bin/ && /usr/local/s-ui/sui migrate && /usr/local/s-ui/sui"]
