package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	masterresolver "github.com/flant/redis-sentinel-proxy/pkg/master_resolver"
	"github.com/flant/redis-sentinel-proxy/pkg/proxy"
	"golang.org/x/sync/errgroup"
)

func main() {
	var (
		localAddr            = ":9999"
		sentinelAddr         = ":26379"
		masterName           = "mymaster"
		masterResolveRetries = 3
	)

	flag.StringVar(&localAddr, "listen", localAddr, "local address")
	flag.StringVar(&sentinelAddr, "sentinel", sentinelAddr, "remote address")
	flag.StringVar(&masterName, "master", masterName, "name of the master redis node")
	flag.IntVar(&masterResolveRetries, "resolve-retries", masterResolveRetries, "number of consecutive retries of the redis master node resolve")
	flag.Parse()

	if err := runProxying(localAddr, sentinelAddr, masterName, masterResolveRetries); err != nil {
		log.Fatalf("Fatal: %s", err)
	}
	log.Println("Exiting...")
}

func runProxying(localAddr, sentinelAddr, masterName string, masterResolveRetries int) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	laddr := resolveTCPAddr(localAddr)
	saddr := resolveTCPAddr(sentinelAddr)

	masterAddrResolver := masterresolver.NewRedisMasterResolver(masterName, saddr, masterResolveRetries)
	rsp := proxy.NewRedisSentinelProxy(laddr, masterAddrResolver)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return masterAddrResolver.UpdateMasterAddressLoop(ctx) })
	eg.Go(func() error { return rsp.Run(ctx) })
	return eg.Wait()
}

func resolveTCPAddr(addr string) *net.TCPAddr {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("Fatal - Failed resolving tcp address: %s", err)
	}
	return tcpAddr
}
