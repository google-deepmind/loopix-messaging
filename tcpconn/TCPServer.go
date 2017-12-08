package tcpconn

import (
	"fmt"
	"net"
	"os"
)

type Server interface {
	ReceivedPacket()
	SendPacket()
	Start()
}

type TCPSocketServer struct {
	Host string
	Port string
}

func (t *TCPSocketServer) ReceivedPacket(packet string) {
	fmt.Println(packet)
}

func (t *TCPSocketServer) SendPacket(packet string) string{
	return packet
}

func (t *TCPSocketServer) Start() {

	addr, err := net.ResolveTCPAddr("tcp", t.Host+":"+t.Port)
	l, err := net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + t.Host + ":" + t.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)
	if err != nil {
		fmt.Println()
	}
	fmt.Println("ReqLen: ", reqLen)
	fmt.Println("ReqLen: ", string(buff[:reqLen]))
	conn.Write([]byte("Message received.\n"))
	conn.Close()
}

