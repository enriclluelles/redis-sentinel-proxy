redis-sentinel-proxy
====================

Small command utility that:

* Given a list of redis sentinel servers listening on `SENTINEL_PORT`, keeps asking it for the address of a master named `NAME`

* Proxies all tcp requests that it receives on `PORT` to that master


Usage:

`./redis-sentinel-proxy -listen IP:PORT -sentinel :SENTINEL_PORT -master NAME`
