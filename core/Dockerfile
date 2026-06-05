FROM debian:bookworm-slim
LABEL maintainer="Alireza <alireza7@gmail.com>"
ENV TZ=Asia/Tehran
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    bash \
    && rm -rf /var/lib/apt/lists/*
COPY sing-box .
COPY runSingbox.sh .
ENTRYPOINT [ "./runSingbox.sh" ]