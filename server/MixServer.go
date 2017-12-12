package server

import (
	node "anonymous-messaging/node"
	"fmt"
	"net"
	"os"
	"anonymous-messaging/packet_format"
)

type MixServer struct {
	Id string
	Host string
	Port string
	mixWorker node.Mix
	listener *net.TCPListener
}

func (m MixServer) ReceivedPacket(packet packet_format.Packet) {
	fmt.Println("> Received packet")

	c := make(chan packet_format.Packet)
	go m.mixWorker.ProcessPacket(packet, c)
	dePacket := <- c

	fmt.Println("> Decoded packet: ", dePacket)
	m.ForwardPacket(dePacket)
}

func (m MixServer) ForwardPacket(packet packet_format.Packet) {
	fmt.Println("> Forwarding packet", packet)

	m.SendPacket(packet, "127.0.0.1", "3332")
}

func (m MixServer) SendPacket(packet packet_format.Packet, host, port string){

	conn, err := net.Dial("tcp", host + ":" + port)
	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write([]byte(packet_format.ToString(packet)))
	defer conn.Close()
}

func (m MixServer) Start() {

	defer m.listener.Close()

	fmt.Println("Listening on " + m.Host + ":" + m.Port)
	m.listenForConnections()
}

func (m MixServer) listenForConnections() {
	for {
		conn, err := m.listener.Accept()

		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}

		go m.handleConnection(conn)
	}
}

func (m MixServer) handleConnection(conn net.Conn) {
	fmt.Println("> Handle Connection")

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println()
	}

	m.ReceivedPacket(packet_format.FromString(string(buff[:reqLen])))
	conn.Close()
}

func NewMixServer(id, host, port string, pubKey, prvKey int) MixServer {
	mixServer := MixServer{}
	mixServer.Id = id
	mixServer.Host = host
	mixServer.Port = port
	mixServer.mixWorker = node.NewMix(id, pubKey, prvKey)

	addr, err := net.ResolveTCPAddr("tcp", mixServer.Host + ":" + mixServer.Port)
	mixServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	return mixServer
}