FROM golang:1.20

WORKDIR /src
COPY  go.mod go.sum /src/
RUN go mod download

COPY redis_sentinel_proxy_test.go /src/

ENTRYPOINT ["/bin/bash", "-c"]
