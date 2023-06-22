redis-sentinel-proxy
====================

Small command utility that:

* Given a redis sentinel server listening on `SENTINEL_PORT`, keeps asking it for the address of a master named `NAME`

* Proxies all tcp requests that it receives on `PORT` to that master

Usage:

`./redis-sentinel-proxy -listen IP:PORT -sentinel :SENTINEL_PORT -master NAME --resolve-retries 10`

testing
============
- install `docker` and `docker-compose`.
- run `make test`