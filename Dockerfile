# ========================================================
# Phase 1: Minimal and Secure Runtime Environment for s-ui
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

# Copy pre-compiled s-ui binary from host backend folder
COPY backend/sui /usr/local/s-ui/sui

# Ensure database, cert and bin directory structures exist with proper permissions
RUN mkdir -p /usr/local/s-ui/db /usr/local/s-ui/bin /usr/local/s-ui/cert

# Copy runSingbox.sh and sing-box binary into bin directory
COPY core/runSingbox.sh /usr/local/s-ui/bin/runSingbox.sh
COPY core/sing-box /usr/local/s-ui/bin/sing-box

RUN chmod +x /usr/local/s-ui/sui /usr/local/s-ui/bin/runSingbox.sh /usr/local/s-ui/bin/sing-box

# Expose s-ui standard dashboard administration port
EXPOSE 2095

# Command to execute migrations first, then run panel main process
CMD ["/bin/sh", "-c", "/usr/local/s-ui/sui migrate && /usr/local/s-ui/sui"]