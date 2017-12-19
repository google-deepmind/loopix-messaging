package server

import (
	node "anonymous-messaging/node"
	"fmt"
	"net"
	"os"
	"anonymous-messaging/packet_format"
	"anonymous-messaging/pki"
)

type MixServer struct {
	Id string
	Host string
	Port string

	mixWorker node.Mix

	listener *net.TCPListener

}

func (m *MixServer) ReceivedPacket(packet packet_format.Packet) {
	fmt.Println("> Received packet")

	c := make(chan packet_format.Packet)
	go m.mixWorker.ProcessPacket(packet, c)
	dePacket := <- c

	fmt.Println("> Decoded packet: ", dePacket)

	if dePacket.Steps[m.Id].Meta.FinalFlag == false{
		m.ForwardPacket(dePacket)
	}
}

func (m *MixServer) ForwardPacket(packet packet_format.Packet) {
	fmt.Println("> Forwarding packet", packet)
	next := packet.Steps[m.Id].Meta
	fmt.Println(next)
	nextHost := next.NextHopHost
	nextPort := next.NextHopPort
	fmt.Println("NEXT PORT: ", nextPort)
	m.SendPacket(packet, nextHost, nextPort)
}

func (m *MixServer) SendPacket(packet packet_format.Packet, host, port string){

	conn, err := net.Dial("tcp", host + ":" + port)
	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write([]byte(packet_format.ToString(packet)))
	defer conn.Close()
}

func (m *MixServer) Start() {
	defer m.Run()

	// m.PublishPublicInfo()

}


func (m *MixServer) Run() {

	fmt.Println("> The Mixserver is running")

	defer m.listener.Close()
	finish := make(chan bool)

	go func() {
		fmt.Println("Listening on " + m.Host + ":" + m.Port)
		m.listenForIncomingConnections()
	}()

	<-finish
}

func (m *MixServer) listenForIncomingConnections(){
	for {
		conn, err := m.listener.Accept()

		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		fmt.Println("Received connection from : ", conn.RemoteAddr())
		go m.handleConnection(conn)
	}
}

func (m *MixServer) handleConnection(conn net.Conn) {
	fmt.Println("> Handle Connection")

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println()
	}

	m.ReceivedPacket(packet_format.FromString(string(buff[:reqLen])))
	conn.Close()
}

func SaveInPKI(m *MixServer) {
	fmt.Println("> Saving into Database")

	db := pki.CreateAndOpenDatabase("./pki/database.db", "./pki/database.db", "sqlite3")

	params := make(map[string]string)
	params["MixId"] = "TEXT"
	params["Host"] = "TEXT"
	params["Port"] = "TEXT"
	params["PubKey"] = "NUM"
	pki.CreateTable(db, "Mixes", params)

	pubInfo := make(map[string]interface{})
	pubInfo["MixId"] = m.Id
	pubInfo["Host"] = m.Host
	pubInfo["Port"] = m.Port
	pubInfo["PubKey"] = m.mixWorker.PubKey
	pki.InsertToTable(db, "Mixes", pubInfo)

	fmt.Println("> Public info of the mixserver saved in database")
}

func NewMixServer(id, host, port string, pubKey, prvKey int) MixServer {
	mixServer := MixServer{}
	mixServer.Id = id
	mixServer.Host = host
	mixServer.Port = port
	mixServer.mixWorker = node.NewMix(id, pubKey, prvKey)

	SaveInPKI(&mixServer)

	addr, err := net.ResolveTCPAddr("tcp", mixServer.Host + ":" + mixServer.Port)
	mixServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	return mixServer
}