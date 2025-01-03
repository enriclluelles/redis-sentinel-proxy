#!/bin/sh

cat <<EOF | tee /redis/sentinel.conf > /dev/null
port 26379

dir /tmp

sentinel resolve-hostnames yes
sentinel monitor redismaster redis-master 6379 2
sentinel down-after-milliseconds redismaster 1000
sentinel parallel-syncs redismaster 1
sentinel failover-timeout redismaster 1000
EOF

redis-server /redis/sentinel.conf --sentinel