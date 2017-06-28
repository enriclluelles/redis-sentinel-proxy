redis-sentinel-proxy
====================

Small command utility that:

* Given a redis sentinel server listening on `SENTINEL_PORT`, keeps asking it for the address of a master named `NAME`

* Proxies all tcp requests that it receives on `PORT` to that master


Usage:

`./redis-sentinel-proxy -listen IP:PORT -sentinel :SENTINEL_PORT -master NAME`

## Usage

Edit `kubernetes/redis-sentinel-proxy-deployment.yaml`:

```bash
vim kubernetes/redis-sentinel-proxy-deployment.yaml
...
        args:
          - "-master"
          - "primary"
          - "-sentinel"
          - "redis-sentinel.$(NAMESPACE):26379" # change this to the sentinel address
```

Create `redis-sentinel-proxy-deployment` that uses `redis-sentinel-proxy`:

```bash
kubectl apply -f kubernetes/redis-sentinel-proxy-deployment.yaml
deployment "redis-sentinel-proxy" configured
```

Check if deployment is running: 

```bash
kubectl get pods
redis-sentinel-proxy-2064359825-s4n0k   1/1       Running   0          1d
```

Expose `redis-sentinel-proxy-deployment`:

```bash
kubectl apply -f kubernetes/redis-sentinel-proxy-service.yaml
```

