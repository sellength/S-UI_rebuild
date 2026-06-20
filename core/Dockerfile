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
    mv /tmp/sing-box-${SINGBOX_VER}-linux-${ARCH_NAME}/sing-box ./sing-box && \
    rm -rf /tmp/sing-box*

COPY runSingbox.sh .
RUN chmod +x sing-box runSingbox.sh

ENTRYPOINT [ "./runSingbox.sh" ]