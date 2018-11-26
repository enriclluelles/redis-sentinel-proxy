FROM alpine:3.4
MAINTAINER Andrey Kolashtov <andrey.kolashtov@flant.com>

RUN go build . -o redis-sentinel-proxy
COPY redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
