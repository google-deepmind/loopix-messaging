/*
	Package server implements the mix server.
*/
package server

import (
	"fmt"
	"net"
	"os"

	"anonymous-messaging/networker"
	"anonymous-messaging/node"
	sphinx "anonymous-messaging/new_packet_format"
	"anonymous-messaging/pki"
	"anonymous-messaging/publics"
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

func (m *MixServer) ReceivedPacket(packet sphinx.SphinxPacket) {
	fmt.Println("> Received packet")

	c := make(chan sphinx.SphinxPacket)
	cHop := make(chan sphinx.Hop)

	go m.ProcessPacket(packet, c, cHop)
	dePacket := <-c
	nextHop := <- cHop

	m.ForwardPacket(nextHop, dePacket)
}

func (m *MixServer) ForwardPacket(nextHop sphinx.Hop, packet sphinx.SphinxPacket) {
	m.SendPacket(packet, nextHop.Address)
}

func (m *MixServer) SendPacket(packet sphinx.SphinxPacket, address string) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write(packet.Bytes())
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

	m.ReceivedPacket(sphinx.PacketFromBytes(buff[:reqLen]))
	conn.Close()
}

func SaveInPKI(m *MixServer, pkiPath string) {
	fmt.Println("> Saving into Database")

	db := pki.OpenDatabase(pkiPath, "sqlite3")

	params := make(map[string]string)
	params["MixId"] = "TEXT"
	params["Host"] = "TEXT"
	params["Port"] = "TEXT"
	params["PubKey"] = "BLOB"
	pki.CreateTable(db, "Mixes", params)

	pubInfo := make(map[string]interface{})
	pubInfo["MixId"] = m.Id
	pubInfo["Host"] = m.Host
	pubInfo["Port"] = m.Port
	pubInfo["PubKey"] = m.PubKey.Bytes()
	pki.InsertToTable(db, "Mixes", pubInfo)

	fmt.Println("> Public info of the mixserver saved in database")

	db.Close()
}

func NewMixServer(id, host, port string, pubKey publics.PublicKey, prvKey publics.PrivateKey, pkiPath string) *MixServer {
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
