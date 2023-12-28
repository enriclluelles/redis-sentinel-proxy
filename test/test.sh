#!/usr/bin/env bash

set -Eeuo pipefail

function stop(){
    docker-compose down > dev/null 2>&1
}
trap stop ERR SIGINT SIGTERM SIGHUP SIGQUIT

{ docker-compose up --build --exit-code-from redis-sentinel-proxy-tests && echo "tests successful"; } || echo "tests failed"
