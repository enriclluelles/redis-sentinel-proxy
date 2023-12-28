package proxy

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/flant/redis-sentinel-proxy/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type masterResolver interface {
	MasterAddress() string
}

type RedisSentinelProxy struct {
	localAddr      *net.TCPAddr
	masterResolver masterResolver
}

func NewRedisSentinelProxy(localAddr *net.TCPAddr, mResolver masterResolver) *RedisSentinelProxy {
	return &RedisSentinelProxy{
		localAddr:      localAddr,
		masterResolver: mResolver,
	}
}

func (r *RedisSentinelProxy) Run(bigCtx context.Context) error {
	listener, err := net.ListenTCP("tcp", r.localAddr)
	if err != nil {
		return err
	}

	errGr, ctx := errgroup.WithContext(bigCtx)
	errGr.Go(func() error { return r.runListenLoop(ctx, listener) })
	errGr.Go(func() error { return closeListenerByContext(ctx, listener) })

	return errGr.Wait()
}

func (r *RedisSentinelProxy) runListenLoop(ctx context.Context, listener *net.TCPListener) error {
	log.Println("Waiting for connections...")
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		go r.proxy(conn)
	}
}

func (r *RedisSentinelProxy) proxy(incoming io.ReadWriteCloser) {
	remote, err := utils.TCPConnectWithTimeout(r.masterResolver.MasterAddress())
	if err != nil {
		defer incoming.Close()
		log.Printf("Error connecting to master: %s", err)
		return
	}

	sigChan := make(chan struct{})
	defer close(sigChan)

	go pipe(incoming, remote, sigChan)
	go pipe(remote, incoming, sigChan)

	<-sigChan
	<-sigChan
}

func closeListenerByContext(ctx context.Context, listener *net.TCPListener) error {
	defer listener.Close()
	<-ctx.Done()
	return nil
}

func pipe(w io.WriteCloser, r io.Reader, sigChan chan<- struct{}) {
	defer func() { sigChan <- struct{}{} }()
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		log.Printf("Error writing content: %s", err)
	}
}
