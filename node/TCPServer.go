package anonymous_messaging

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

type TCPServer struct {
	Host string
	//IP string
	Port string
}

func (t *TCPServer) ReceivedPacket(packet string) {
	fmt.Println(packet)
}

func (t *TCPServer) SendPacket(packet string) string{
	return packet
}

func (t *TCPServer) Start() {
	l, err := net.Listen("tcp", t.Host+":"+t.Port)
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
	conn.Write([]byte("Message received.\n"))
	conn.Close()
}

