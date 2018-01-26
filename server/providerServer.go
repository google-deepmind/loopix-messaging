package server

import (
	"anonymous-messaging/node"
	"net"
	"anonymous-messaging/networker"
	"os"
	"github.com/glog"
	"fmt"
	"bytes"
	"anonymous-messaging/pki"
	"anonymous-messaging/publics"
	"io/ioutil"
	"anonymous-messaging/helpers"
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

	assignedClients map[string]ClientRecord

	Config publics.MixPubs
}

type ClientRecord struct {
	Id string
	Host string
	Port string
	PubKey []byte
	Token []byte
}

func (p *ProviderServer) ReceivedPacket(packet []byte) {
	fmt.Println("> Provider received packet")

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

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		glog.Info("Connection handling failed")
	}

	p.ReceivedPacket(buff[:reqLen])
	conn.Close()
}

func (p *ProviderServer) StoreMessage(message []byte, inboxId string, messageId string) {

	path := fmt.Sprintf("./inboxes/%s", inboxId)

	fmt.Println(path)
	fileName := path + "/" + messageId
	fmt.Println(fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic("Unable to create directory for storage file! - " + err.Error())
		}

	}
	file, err := os.Create(fileName)
	fmt.Println("Created path")
	fmt.Println(file)

	if err != nil {
		glog.Error("Provider error while storing message: ", err.Error())
	}
	defer file.Close()
	file.Write(message)

}

func (p *ProviderServer) AuthenticateUser(clientId string, clientToken []byte) bool{

	if bytes.Compare(p.assignedClients[clientId].Token, clientToken) == 0 {
		return true
	}
	return false
}

func (p *ProviderServer) FetchMessages(clientId string) error{

	path := fmt.Sprintf("./inboxes/%s", clientId)

	_, err := os.Stat(path)
	if err != nil{
		return err
	}

	files, err := ioutil.ReadDir(path)

	for _, f := range files {
		fmt.Println(f.Name())
		dat, err := ioutil.ReadFile(path + "/" + f.Name())
		if err !=nil {
			return err
		}
		fmt.Println(dat)

		address := p.assignedClients[clientId].Host + ":" + p.assignedClients[clientId].Port
		fmt.Println("ADR: ", address)
	//	p.SendPacket(dat, address)
	}
	return nil
}

func (p *ProviderServer) SaveInPKI(path string) {

	db := pki.OpenDatabase(path, "sqlite3")

	params := make(map[string]string)
	params["Id"] = "TEXT"
	params["Typ"] = "TEXT"
	params["Config"] = "BLOB"
	pki.CreateTable(db, "Providers", params)

	pubsBytes, err := publics.MixPubsToBytes(p.Config)
	if err != nil {
		panic(err)
	}

	pki.InsertIntoTable(db, "Providers", p.Id, "Provider", pubsBytes)
	db.Close()
	fmt.Println("> Provider public information saved in the database")
}

func (p *ProviderServer) Start() {
	defer p.Run()
}

func (p *ProviderServer) Run() {

	fmt.Println("> The provider server is running")

	defer p.listener.Close()
	finish := make(chan bool)

	go func() {
		fmt.Println("Listening on " + p.Host + ":" + p.Port)
		p.ListenForIncomingConnections()
	}()

	<-finish
}

func NewProviderServer(id string, host string, port string, pubKey []byte, prvKey []byte, pkiPath string) *ProviderServer {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	providerServer := ProviderServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	providerServer.Config = publics.MixPubs{Id: providerServer.Id, Host: providerServer.Host, Port: providerServer.Port, PubKey: providerServer.PubKey}
	providerServer.SaveInPKI(pkiPath)

	addr, err := helpers.ResolveTCPAddress(providerServer.Host, providerServer.Port)

	if err != nil {
		panic(err)
	}
	providerServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	return &providerServer
}