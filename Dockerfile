FROM golang:1.20 AS builder
LABEL Andrey Kolashtov <andrey.kolashtov@flant.com>

WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download

COPY main.go Makefile /src/
COPY pkg /src/pkg

RUN make build CGO_ENABLED=0 GOOS=linux GOARCH=amd64

FROM alpine:3.17

COPY --from=builder /src/bin/redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy
RUN apk --update --no-cache add redis

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy"]
CMD ["-master", "mymaster"]
