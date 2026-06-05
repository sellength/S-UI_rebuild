# ========================================================
# Phase 1: Minimal and Secure Runtime Environment for s-ui
# ========================================================
FROM debian:bookworm-slim

LABEL org.opencontainers.image.authors="coolnnek" \
      org.opencontainers.image.description="s-ui: Advanced Dual-Stack Sing-box User Interface Dashboard"

# Install essential runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    sqlite3 \
    curl \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Set system default timezone
ENV TZ=Asia/Shanghai

WORKDIR /usr/local/s-ui

# Copy pre-compiled s-ui binary from host backend folder
COPY backend/sui /usr/local/s-ui/sui

# Ensure database, cert and bin directory structures exist with proper permissions
RUN mkdir -p /usr/local/s-ui/db /usr/local/s-ui/bin /usr/local/s-ui/cert && \
    chmod +x /usr/local/s-ui/sui

# Expose s-ui standard dashboard administration port
EXPOSE 2095

# Command to execute migrations first, then run panel main process
CMD ["/bin/sh", "-c", "/usr/local/s-ui/sui migrate && /usr/local/s-ui/sui"]