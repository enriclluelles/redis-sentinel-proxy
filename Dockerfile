FROM alpine:3.4
MAINTAINER Anubhav Mishra <anubhavmishra@me.com>

# copy binary
COPY redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
