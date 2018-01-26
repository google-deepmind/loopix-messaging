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
	"anonymous-messaging/pki"
	"anonymous-messaging/publics"
	"anonymous-messaging/helpers"
)

type MixServerIt interface {
	networker.NetworkServer
	networker.NetworkClient
}

type MixServer struct {
	Id   string
	Host string
	Port string
	listener *net.TCPListener
	node.Mix

	Config publics.MixPubs
}

func (m *MixServer) ReceivedPacket(packet []byte) {
	fmt.Println("> Mix Received packet: ", packet)

	c := make(chan []byte)
	cAdr := make(chan string)
	cFlag := make(chan string)

	go m.ProcessPacket(packet, c, cAdr, cFlag)
	dePacket := <-c
	nextHopAdr := <- cAdr
	flag := <- cFlag

	if flag == "\xF1" {
		m.ForwardPacket(dePacket, nextHopAdr)
	} else  {
		fmt.Println("Not forwarding")
	}
}

func (m *MixServer) ForwardPacket(packet []byte, address string) {
	m.SendPacket(packet, address)
}

func (m *MixServer) SendPacket(packet []byte, address string) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write(packet)
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

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println("Connection Handle failed")
	}

	m.ReceivedPacket(buff[:reqLen])
	conn.Close()
}

func (m *MixServer) SaveInPKI(pkiPath string) {
	fmt.Println("> Saving into Database")

	db := pki.OpenDatabase(pkiPath, "sqlite3")

	params := make(map[string]string)
	params["Id"] = "TEXT"
	params["Typ"] = "TEXT"
	params["Config"] = "BLOB"
	pki.CreateTable(db, "Mixes", params)

	configBytes, err := publics.MixPubsToBytes(m.Config)
	if err != nil {
		panic(err)
	}

	pki.InsertIntoTable(db, "Mixes", m.Id, "Mix", configBytes)

	fmt.Println("> Public info of the mixserver saved in database")
	db.Close()
}

func NewMixServer(id, host, port string, pubKey []byte, prvKey []byte, pkiPath string) *MixServer {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	mixServer := MixServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	mixServer.Config = publics.MixPubs{Id : mixServer.Id, Host: mixServer.Host, Port: mixServer.Port, PubKey: mixServer.PubKey}

	mixServer.SaveInPKI(pkiPath)

	addr, err := helpers.ResolveTCPAddress(mixServer.Host, mixServer.Port)

	if err != nil {
		panic(err)
	}
	mixServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	return &mixServer
}
