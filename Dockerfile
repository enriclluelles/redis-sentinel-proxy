FROM golang:1.14.3-alpine AS build

COPY main.go /src/redis-sentinel-proxy/

WORKDIR /src/redis-sentinel-proxy/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

FROM alpine:3.4
MAINTAINER Anubhav Mishra <anubhavmishra@me.com>

# copy binary
COPY --from=build /src/redis-sentinel-proxy/redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
