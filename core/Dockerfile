# ==========================================
# Phase 1: Build Sing-Box with tags
# ==========================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG SINGBOX_VER=1.13.13
RUN apk add --no-cache git
RUN git clone --branch v${SINGBOX_VER} --depth 1 https://github.com/sagernet/sing-box.git /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags with_v2ray_api -ldflags "-s -w" -o /out/sing-box ./cmd/sing-box

# ==========================================
# Phase 2: Final Runtime Image
# ==========================================
FROM alpine:latest
LABEL maintainer="sellength"
ENV TZ=Asia/Shanghai
WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash \
    curl \
    tar

COPY --from=builder /out/sing-box ./sing-box
COPY runSingbox.sh .
RUN chmod +x sing-box runSingbox.sh

ENTRYPOINT [ "./runSingbox.sh" ]