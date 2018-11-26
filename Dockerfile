FROM golang:1.10
MAINTAINER Andrey Kolashtov <andrey.kolashtov@flant.com>
RUN mkdir /redis-sentinel-proxy
ADD . /redis-sentinel-proxy/
WORKDIR /redis-sentinel-proxy
RUN apt-get update
RUN apt-get install redis-tools
RUN go build -o redis-sentinel-proxy .
RUN mv /redis-sentinel-proxy/redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
