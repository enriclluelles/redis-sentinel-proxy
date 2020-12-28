package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

var (
	masterAddr *net.TCPAddr
	raddr      *net.TCPAddr
	saddr      *net.TCPAddr

	localAddr    = flag.String("listen", ":9999", "local address")
	sentinelAddr = flag.String("sentinel", ":26379", "remote address")
	masterName   = flag.String("master", "", "name of the master redis node")
	password     = flag.String("password", "", "password (if any) to authenticate")
	debug   	 = flag.Bool("debug", false, "sets debug mode")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		log.Fatal("Failed to resolve local address: %s", err)
	}
	saddr, err = net.ResolveTCPAddr("tcp", *sentinelAddr)
	if err != nil {
		log.Fatal("Failed to resolve sentinel address: %s", err)
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
	var err error
	for {
		masterAddr, err = getMasterAddr(saddr, *masterName, *password)
		if err != nil {
			log.Println(err)
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

func getMasterAddr(sentinelAddress *net.TCPAddr, masterName string, password string) (*net.TCPAddr, error) {
	conn, err := net.DialTCP("tcp", nil, sentinelAddress)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if len(password) > 0 {
		conn.Write([]byte(fmt.Sprintf("AUTH %s\n", password)))
		if *debug {
			fmt.Println("> AUTH ", password)
		}
		authResp := make([]byte, 256)
		_, err = conn.Read(authResp)

		if *debug {
			fmt.Println("< ", string(authResp))
		}
	}

	if *debug {
		fmt.Println("> sentinel get-master-addr-by-name ", masterName)
	}
	conn.Write([]byte(fmt.Sprintf("sentinel get-master-addr-by-name %s\n", masterName)))

	b := make([]byte, 256)
	_, err = conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(string(b), "\r\n")
	if *debug {
		fmt.Println("< ", string(b))
	}

	if len(parts) < 5 {
		err = errors.New(fmt.Sprintf("Couldn't get master address from sentinel: %s", string(b)))
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
