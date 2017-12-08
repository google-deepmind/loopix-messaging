package server

import (
	node "anonymous-messaging/node"
	"fmt"
	"net"
	"os"
)

type MixServer struct {
	Id string
	Host string
	Port string
	mixWorker node.Mix
}

func (m MixServer) ReceivedPacket(packet string) {
	m.mixWorker.ProcessPacket(packet)
	//fmt.Println(packet)
}

func (m MixServer) SendPacket(packet string){
	fmt.Println(packet)
}

func (m MixServer) Start() {

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
		//fmt.Println(conn)
		go m.handleConnection(conn)
	}
}

func (m MixServer) handleConnection(conn net.Conn) {
	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println()
	}

	m.mixWorker.ProcessPacket(string(buff[:reqLen]))

	conn.Write([]byte("Message received.\n"))
	conn.Close()
}

func NewMixServer(id, host, port string, pubKey, prvKey int) MixServer {
	mixServer := MixServer{}
	mixServer.Id = id
	mixServer.Host = host
	mixServer.Port = port
	mixServer.mixWorker = node.NewMix(id, pubKey, prvKey)
	return mixServer
}