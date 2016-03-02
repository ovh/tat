FROM alpine
MAINTAINER yvonnick.esnault@corp.ovh.com
RUN apk add --update -t deps wget ca-certificates && \
    mkdir /app && cd /app && \
    TAT_VERSION=1.1.0 && \
    wget https://github.com/ovh/tat/releases/download/v${TAT_VERSION}/tat-linux-amd64 && \
    chmod +x tat-linux-amd64 && \
    chown -R nobody:nogroup /app && \
    apk del --purge deps
USER nobody
CMD ["/app/tat-linux-amd64"]
