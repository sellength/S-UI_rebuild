FROM alpine:latest
LABEL maintainer="sellength"
ENV TZ=Asia/Tehran
WORKDIR /app
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash
COPY sing-box .
COPY runSingbox.sh .
ENTRYPOINT [ "./runSingbox.sh" ]