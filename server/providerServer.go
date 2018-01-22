package server

import (
	"anonymous-messaging/node"
	"net"
	"anonymous-messaging/networker"
	"os"
	"github.com/glog"
	"anonymous-messaging/publics"
	"fmt"
	"anonymous-messaging/pki"
	"flag"
)

type ProviderIt interface {
	networker.NetworkServer
	networker.NetworkClient
}

type ProviderServer struct {
	Id string
	Host string
	Port string
	node.Mix
	listener *net.TCPListener
}

func (p *ProviderServer) ReceivedPacket(packet []byte) {

	c := make(chan []byte)
	cAdr := make(chan string)
	cFlag := make(chan string)

	go p.ProcessPacket(packet, c, cAdr, cFlag)
	dePacket := <-c
	nextHop := <- cAdr
	flag := <- cFlag

	switch flag {
	case "\xF1":
		p.ForwardPacket(dePacket, nextHop)
	case "\xF0":
		p.StoreMessage(dePacket, nextHop, "TMP_MESSAGE_ID")
	}
}

func (p *ProviderServer) ForwardPacket(packet []byte, address string){
	p.SendPacket(packet, address)
}

func (p *ProviderServer) SendPacket(packet []byte, address string) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		glog.Info("Error in Provider Send Packet:  ", err.Error())
		os.Exit(1)
	}

	conn.Write(packet)
	defer conn.Close()
}


func (p *ProviderServer) ListenForIncomingConnections() {
	for {
		conn, err := p.listener.Accept()

		if err != nil {
			glog.Info("Error in Provider connection accepting:  ", err.Error())
			os.Exit(1)
		}
		glog.Info("Provider received connection from: ", conn.RemoteAddr())
		go p.HandleConnection(conn)
	}
}

func (p *ProviderServer) HandleConnection(conn net.Conn) {
	glog.Info("Provider handle connection")

	var buff []byte
	reqLen, err := conn.Read(buff)

	if err != nil {
		glog.Info("Connection handling failed")
	}

	p.ReceivedPacket(buff[:reqLen])
	conn.Close()
}

func (p *ProviderServer) StoreMessage(message []byte, inboxId string, messageId string) {

	path := fmt.Sprintf("./inboxes/%s", inboxId)
	fileName := path + "/" + messageId
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic("Unable to create directory for storage file! - " + err.Error())
		}
	}
	file, err := os.Create(fileName)

	if err != nil {
		glog.Error("Provider error while storing message: ", err.Error())
	}
	defer file.Close()
	file.Write(message)
}

func (p *ProviderServer) AuthenticateUser() {}


func (p *ProviderServer) SaveInPKI(path string) {

	db := pki.OpenDatabase(path, "sqlite3")

	params := make(map[string]string)
	params["ProviderId"] = "TEXT"
	params["Host"] = "TEXT"
	params["Port"] = "TEXT"
	params["PubKey"] = "BLOB"
	pki.CreateTable(db, "Providers", params)

	pubInfo := make(map[string]interface{})
	pubInfo["ProviderId"] = p.Id
	pubInfo["Host"] = p.Host
	pubInfo["Port"] = p.Port
	pubInfo["PubKey"] = p.PubKey.Bytes()
	pki.InsertToTable(db, "Providers", pubInfo)

	flag.Parse()
	glog.Info("Provider info stored in the PKI database")

	db.Close()
}


func NewProviderServer(id string, host string, port string, pubKey publics.PublicKey, prvKey publics.PrivateKey, pkiPath string) *ProviderServer {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	providerServer := ProviderServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	providerServer.SaveInPKI(pkiPath)

	addr, err := net.ResolveTCPAddr("tcp", providerServer.Host+":"+providerServer.Port)
	providerServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	return &providerServer
}