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
	"anonymous-messaging/config"
	"anonymous-messaging/helpers"
	"log"
	"anonymous-messaging/logging"
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

	Config config.MixPubs

	infoLogger *log.Logger
	errorLogger *log.Logger
}

func (m *MixServer) ReceivedPacket(packet []byte) error{
	m.infoLogger.Println(fmt.Sprintf("%s: Received new packet", m.Id))

	c := make(chan []byte)
	cAdr := make(chan string)
	cFlag := make(chan string)
	errCh := make(chan error)

	go m.ProcessPacket(packet, c, cAdr, cFlag, errCh)
	dePacket := <-c
	nextHopAdr := <- cAdr
	flag := <- cFlag
	err := <- errCh

	if err != nil{
		return err
	}

	if flag == "\xF1" {
		m.ForwardPacket(dePacket, nextHopAdr)
	} else  {
		m.infoLogger.Println(fmt.Sprintf("%s: Packet has non-forward flag", m.Id))
	}
	return nil
}

func (m *MixServer) ForwardPacket(packet []byte, address string) {
	m.SendPacket(packet, address)
}

func (m *MixServer) SendPacket(packet []byte, address string) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		m.errorLogger.Println(err)
		os.Exit(1)
	}

	conn.Write(packet)
	defer conn.Close()
}

func (m *MixServer) Start() {
	defer m.Run()

	f, err := os.OpenFile("./logging/network_logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755)

	if err != nil{
		panic(err)
	}
	// defer f.Close()

	m.infoLogger = logging.NewInitLogger(f)
	m.errorLogger = logging.NewErrorLogger(f)
}

func (m *MixServer) Run() {

	defer m.listener.Close()
	finish := make(chan bool)

	go func() {
		m.infoLogger.Println(fmt.Sprintf("%s: Listening on %s", m.Id, m.Host + ":" + m.Port))
		m.ListenForIncomingConnections()
	}()

	<-finish
}

func (m *MixServer) ListenForIncomingConnections() {
	for {
		conn, err := m.listener.Accept()

		if err != nil {
			m.errorLogger.Println(err)
			os.Exit(1)
		}
		m.infoLogger.Println(fmt.Sprintf("%s: Received connection from %s", m.Id, conn.RemoteAddr()))
		go m.HandleConnection(conn)
	}
}

func (m *MixServer) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		m.errorLogger.Println(err)
	}

	err = m.ReceivedPacket(buff[:reqLen])
	if err != nil{
		m.errorLogger.Println(err)
	}
	conn.Close()
}

func (m *MixServer) SaveInPKI(pkiPath string) error {

	db, err := pki.OpenDatabase(pkiPath, "sqlite3")
	defer db.Close()

	if err != nil{
		return err
	}

	params := make(map[string]string)
	params["Id"] = "TEXT"
	params["Typ"] = "TEXT"
	params["Config"] = "BLOB"

	pki.CreateTable(db, "Mixes", params)

	configBytes, err := config.MixPubsToBytes(m.Config)
	if err != nil {
		return err
	}

	err = pki.InsertIntoTable(db, "Mixes", m.Id, "Mix", configBytes)
	if err != nil {
		return err
	}

	return nil
}

func NewMixServer(id, host, port string, pubKey []byte, prvKey []byte, pkiPath string) (*MixServer, error) {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	mixServer := MixServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	mixServer.Config = config.MixPubs{Id : mixServer.Id, Host: mixServer.Host, Port: mixServer.Port, PubKey: mixServer.PubKey}

	err := mixServer.SaveInPKI(pkiPath)
	if err != nil {
		return nil, err
	}

	addr, err := helpers.ResolveTCPAddress(mixServer.Host, mixServer.Port)

	if err != nil {
		return nil, err
	}
	mixServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &mixServer, nil
}
