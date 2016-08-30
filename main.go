package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	masterAddr *net.TCPAddr
	saddrs     []*net.TCPAddr

	localAddr     = flag.String("listen", ":9999", "local address")
	sentinelAddrs = flag.String("sentinel", ":26379", "List of remote address, separated by coma")
	masterName    = flag.String("master", "", "name of the master redis node")
	pidFile       = flag.String("pidfile", "", "Location of the pid file")
)

func main() {
	flag.Parse()

	if *pidFile != "" {
		f, err := os.Create(*pidFile)
		if err != nil {
			log.Fatalf("Unable to create pidfile: %s", err)
		}

		fmt.Fprintf(f, "%d\n", os.Getpid())

		f.Close()
	}

	sentinels := strings.Split(*sentinelAddrs, ",")

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		log.Fatal("Failed to resolve local address: %s", err)
	}
	for _, sentinel := range sentinels {
		saddr, err := net.ResolveTCPAddr("tcp", sentinel)
		if err != nil {
			log.Fatal("Failed to resolve sentinel address: %s", err)
		}
		saddrs = append(saddrs, saddr)
	}

	go master()

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		go proxy(conn, masterAddr)
	}
}

func master() {
	for {
		for _, saddr := range saddrs {
			result_channel := make(chan *net.TCPAddr, 1)
			go func(saddr *net.TCPAddr) {
				master, err := getMasterAddr(saddr, *masterName)
				if err != nil {
					log.Println(err)
				}
				result_channel <- master
			}(saddr)
			select {
			case result := <-result_channel:
				masterAddr = result
			case <-time.After(time.Second * 2):
				log.Println("Sentinel timed out")
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func pipe(r io.Reader, w io.WriteCloser) {
	io.Copy(w, r)
	w.Close()
}

func proxy(local io.ReadWriteCloser, remoteAddr *net.TCPAddr) {
	remote, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		log.Println(err)
		local.Close()
		return
	}
	go pipe(local, remote)
	go pipe(remote, local)
}

func getMasterAddr(sentinelAddress *net.TCPAddr, masterName string) (*net.TCPAddr, error) {
	conn, err := net.DialTCP("tcp", nil, sentinelAddress)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("sentinel get-master-addr-by-name %s\n", masterName)))

	b := make([]byte, 256)
	_, err = conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(string(b), "\r\n")

	if len(parts) < 5 {
		err = errors.New("Couldn't get master address from sentinel")
		return nil, err
	}

	//getting the string address for the master node
	stringaddr := fmt.Sprintf("%s:%s", parts[2], parts[4])
	addr, err := net.ResolveTCPAddr("tcp", stringaddr)

	if err != nil {
		return nil, err
	}

	//check that there's actually someone listening on that address
	conn2, err := net.DialTCP("tcp", nil, addr)
	if err == nil {
		defer conn2.Close()
	}

	return addr, err
}
