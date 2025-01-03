package utils

import (
	"net"
	"time"
)

func TCPConnectWithTimeout(addr string) (net.Conn, error) {
	remote, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return nil, err
	}
	return remote, nil
}
