package server

import (
	"anonymous-messaging/networker"
	"anonymous-messaging/node"
	"anonymous-messaging/packet_format"
	"anonymous-messaging/pki"
	"fmt"
	"net"
	"os"
)

type MixServerIt interface {
	networker.NetworkServer
	networker.NetworkClient
}

type MixServer struct {
	Id   string
	Host string
	Port string

	node.Mix

	listener *net.TCPListener
}

func (m *MixServer) ReceivedPacket(packet packet_format.Packet) {
	fmt.Println("> Received packet")

	c := make(chan packet_format.Packet)
	go m.ProcessPacket(packet, c)
	dePacket := <-c

	fmt.Println("> Decoded packet: ", dePacket)

	if dePacket.Steps[m.Id].Meta.FinalFlag {
		m.ForwardPacket(dePacket)
	}
}

func (m *MixServer) ForwardPacket(packet packet_format.Packet) {
	fmt.Println("> Forwarding packet", packet)
	next := packet.Steps[m.Id].Meta

	m.SendPacket(packet, next.NextHopHost, next.NextHopPort)
}

func (m *MixServer) SendPacket(packet packet_format.Packet, host, port string) {

	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write([]byte(packet_format.ToString(packet)))
	defer conn.Close()
}

func (m *MixServer) Start() {
	defer m.Run()
}

func (m *MixServer) Run() {

	fmt.Println("> The Mixserver is running")

	defer m.listener.Close()
	finish := make(chan bool)

	go func() {
		fmt.Println("Listening on " + m.Host + ":" + m.Port)
		m.ListenForIncomingConnections()
	}()

	<-finish
}

func (m *MixServer) ListenForIncomingConnections() {
	for {
		conn, err := m.listener.Accept()

		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		fmt.Println("Received connection from : ", conn.RemoteAddr())
		go m.HandleConnection(conn)
	}
}

func (m *MixServer) HandleConnection(conn net.Conn) {
	fmt.Println("> Handle Connection")

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println("Connection Handle failed")
	}

	m.ReceivedPacket(packet_format.FromString(string(buff[:reqLen])))
	conn.Close()
}

func SaveInPKI(m *MixServer, pkiPath string) {
	fmt.Println("> Saving into Database")

	db := pki.OpenDatabase(pkiPath, "sqlite3")

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
	pubInfo["PubKey"] = m.PubKey
	pki.InsertToTable(db, "Mixes", pubInfo)

	fmt.Println("> Public info of the mixserver saved in database")

	db.Close()
}

func NewMixServer(id, host, port string, pubKey, prvKey int, pkiPath string) *MixServer {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	mixServer := MixServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	SaveInPKI(&mixServer, pkiPath)

	addr, err := net.ResolveTCPAddr("tcp", mixServer.Host+":"+mixServer.Port)
	mixServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	return &mixServer
}
