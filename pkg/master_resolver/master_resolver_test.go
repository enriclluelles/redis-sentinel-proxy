package masterresolver

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func Test_redisMasterFromSentinelAddr(t *testing.T) {
	type args struct {
		sentinelAddress *net.TCPAddr
		masterName      string
	}

	mockServerAddr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12700}
	tests := []struct {
		name    string
		args    args
		want    *net.TCPAddr
		wantErr bool
	}{
		{
			name: "all is ok",
			want: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12700},
			args: args{
				sentinelAddress: mockServerAddr,
				masterName:      "test-master",
			},
		},
		{
			name:    "fail with error",
			wantErr: true,
			args: args{
				sentinelAddress: mockServerAddr,
				masterName:      "bad-master",
			},
		},
	}

	go mockSentinelServer(mockServerAddr)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := redisMasterFromSentinelAddr(tt.args.sentinelAddress, tt.args.masterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("redisMasterFromSentinelAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("redisMasterFromSentinelAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockSentinelServer(addr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil || conn == nil {
			log.Println("could not accept connection")
		}
		log.Println("accepted test connection")
		testAccept(conn, addr)
	}
}

func testAccept(conn net.Conn, addr *net.TCPAddr) {
	defer conn.Close()
	out := make([]byte, 256)
	if _, err := conn.Read(out); err != nil {
		return
	}

	var masterAddr string
	if bytes.HasPrefix(out, []byte("sentinel get-master-addr-by-name test-master")) {
		masterAddr = strings.Join([]string{"tralala", "tralala", addr.IP.String(), "tralala", strconv.Itoa(addr.Port)}, "\r\n")
	} else {
		masterAddr = strings.Join([]string{"tralala", "tralala", addr.IP.String(), "tralala", "40"}, "\r\n")
	}

	if _, err := conn.Write([]byte(masterAddr)); err != nil {
		log.Println("could not write payload to TCP server:", err)
	}
}
