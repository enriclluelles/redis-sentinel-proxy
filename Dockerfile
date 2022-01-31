FROM golang:1.17 AS builder
LABEL Andrey Kolashtov <andrey.kolashtov@flant.com>

ADD . /redis-sentinel-proxy/
WORKDIR /redis-sentinel-proxy
RUN go mod init redis-sentinel-proxy && \
    go build -o redis-sentinel-proxy .

FROM alpine:3.14

COPY --from=builder /redis-sentinel-proxy/redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy
RUN apk --update --no-cache add redis

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
