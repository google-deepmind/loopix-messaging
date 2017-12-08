package tcpconn

import (
	"fmt"
	"net"
	"os"

	node "anonymous-messaging/node"
)

type TCPServer interface {
	ReceivedPacket()
	SendPacket()
	Start()
}

type MixServer struct {
	Id string
	Host string
	Port string
	mixWorker node.Mix
}

func (m *MixServer) ReceivedPacket(packet string) {
	go m.mixWorker.ProcessPacket(packet)
}

func (m *MixServer) SendPacket(packet string) string{
	return packet
}

func (m *MixServer) Start() {

	addr, err := net.ResolveTCPAddr("tcp", m.Host+":"+m.Port)
	l, err := net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening on " + m.Host + ":" + m.Port)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		go m.handleConnection(conn)
	}
}

func (m *MixServer) handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println()
	}

	fmt.Println("ReqLen: ", reqLen)
	fmt.Println("ReqLen: ", string(buff[:reqLen]))
	m.mixWorker.ProcessPacket(string(buff[:reqLen]))

	conn.Write([]byte("Message received.\n"))
	conn.Close()
}

func NewMixServer(id, host, port string, pubkey, prvkey int) MixServer {
	mixServer := MixServer{}
	mixServer.Id = id
	mixServer.Host = host
	mixServer.Port = port
	mixServer.mixWorker = node.Mix{Id : id, PubKey : pubkey, PrvKey : prvkey}
	return mixServer
}
