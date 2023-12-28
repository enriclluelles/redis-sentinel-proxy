package masterresolver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/flant/redis-sentinel-proxy/pkg/utils"
)

type RedisMasterResolver struct {
	masterName               string
	sentinelAddr             *net.TCPAddr
	retryOnMasterResolveFail int

	masterAddrLock           *sync.RWMutex
	initialMasterResolveLock chan struct{}

	masterAddr string
}

func NewRedisMasterResolver(masterName string, sentinelAddr *net.TCPAddr, retryOnMasterResolveFail int) *RedisMasterResolver {
	return &RedisMasterResolver{
		masterName:               masterName,
		sentinelAddr:             sentinelAddr,
		retryOnMasterResolveFail: retryOnMasterResolveFail,
		masterAddrLock:           &sync.RWMutex{},
		initialMasterResolveLock: make(chan struct{}),
	}
}

func (r *RedisMasterResolver) MasterAddress() string {
	<-r.initialMasterResolveLock

	r.masterAddrLock.RLock()
	defer r.masterAddrLock.RUnlock()
	return r.masterAddr
}

func (r *RedisMasterResolver) setMasterAddress(masterAddr *net.TCPAddr) {
	r.masterAddrLock.Lock()
	defer r.masterAddrLock.Unlock()
	r.masterAddr = masterAddr.String()
}

func (r *RedisMasterResolver) updateMasterAddress() error {
	masterAddr, err := redisMasterFromSentinelAddr(r.sentinelAddr, r.masterName)
	if err != nil {
		log.Println(err)
		return err
	}
	r.setMasterAddress(masterAddr)
	return nil
}

func (r *RedisMasterResolver) UpdateMasterAddressLoop(ctx context.Context) error {
	if err := r.initialMasterAdressResolve(); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var err error
	for errCount := 0; errCount <= r.retryOnMasterResolveFail; {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		err = r.updateMasterAddress()
		if err != nil {
			errCount++
		} else {
			errCount = 0
		}
	}
	return err
}

func (r *RedisMasterResolver) initialMasterAdressResolve() error {
	defer close(r.initialMasterResolveLock)
	return r.updateMasterAddress()
}

func redisMasterFromSentinelAddr(sentinelAddress *net.TCPAddr, masterName string) (*net.TCPAddr, error) {
	conn, err := utils.TCPConnectWithTimeout(sentinelAddress.String())
	if err != nil {
		return nil, fmt.Errorf("error connecting to sentinel: %w", err)
	}
	defer conn.Close()

	getMasterCommand := fmt.Sprintf("sentinel get-master-addr-by-name %s\n", masterName)
	if _, err := conn.Write([]byte(getMasterCommand)); err != nil {
		return nil, fmt.Errorf("error writing to sentinel: %w", err)
	}

	b := make([]byte, 256)
	if _, err := conn.Read(b); err != nil {
		return nil, fmt.Errorf("error getting info from sentinel: %w", err)
	}

	parts := strings.Split(string(b), "\r\n")

	if len(parts) < 5 {
		return nil, errors.New("couldn't get master address from sentinel")
	}

	// getting the string address for the master node
	stringaddr := fmt.Sprintf("%s:%s", parts[2], parts[4])
	addr, err := net.ResolveTCPAddr("tcp", stringaddr)
	if err != nil {
		return nil, fmt.Errorf("error resolving redis master: %w", err)
	}

	// check that there's actually someone listening on that address
	if err := checkTCPConnect(addr); err != nil {
		return nil, fmt.Errorf("error checking redis master: %w", err)
	}
	return addr, nil
}

func checkTCPConnect(addr *net.TCPAddr) error {
	conn, err := utils.TCPConnectWithTimeout(addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
