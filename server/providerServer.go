package server

import (
	"anonymous-messaging/node"
	"net"
	"anonymous-messaging/networker"
	"os"
	"fmt"
	"bytes"
	"anonymous-messaging/pki"
	"anonymous-messaging/config"
	"io/ioutil"
	"anonymous-messaging/helpers"
	"anonymous-messaging/logging"
	"log"
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

	Config config.MixPubs
	infoLogger *log.Logger
	errorLogger *log.Logger
}

type ClientRecord struct {
	Id string
	Host string
	Port string
	PubKey []byte
	Token []byte
}

func (p *ProviderServer) ReceivedPacket(packet []byte) error{
	p.infoLogger.Println(fmt.Sprintf("%s: received new packet", p.Id))

	c := make(chan []byte)
	cAdr := make(chan string)
	cFlag := make(chan string)
	errCh := make(chan error)

	go p.ProcessPacket(packet, c, cAdr, cFlag, errCh)
	dePacket := <-c
	nextHop := <- cAdr
	flag := <- cFlag
	err := <- errCh

	if err != nil{
		return err
	}
	switch flag {
	case "\xF1":
		p.ForwardPacket(dePacket, nextHop)
	case "\xF0":
		p.StoreMessage(dePacket, nextHop, "TMP_MESSAGE_ID")
	}
	return nil
}

func (p *ProviderServer) ForwardPacket(packet []byte, address string){
	p.SendPacket(packet, address)
	p.infoLogger.Println(fmt.Sprintf("%s: forwarded packet", p.Id))
}

func (p *ProviderServer) SendPacket(packet []byte, address string) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		p.errorLogger.Println(err)
		os.Exit(1)
	}

	conn.Write(packet)
	defer conn.Close()
}


func (p *ProviderServer) ListenForIncomingConnections() {
	for {
		conn, err := p.listener.Accept()

		if err != nil {
			p.errorLogger.Println(err)
			os.Exit(1)
		}
		p.infoLogger.Println(fmt.Sprintf("%s: Received new connection from %s", p.Id, conn.RemoteAddr()))
		go p.HandleConnection(conn)
	}
}

func (p *ProviderServer) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		p.errorLogger.Println(err)
	}

	err = p.ReceivedPacket(buff[:reqLen])
	if err != nil {
		p.errorLogger.Println(err)
	}
	
	conn.Close()
}

func (p *ProviderServer) StoreMessage(message []byte, inboxId string, messageId string) {

	path := fmt.Sprintf("./inboxes/%s", inboxId)

	fmt.Println(path)
	fileName := path + "/" + messageId
	fmt.Println(fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			p.errorLogger.Println("Unable to create directory for storage file! - " + err.Error())
		}

	}
	file, err := os.Create(fileName)

	if err != nil {
		p.errorLogger.Println(err)
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
		p.infoLogger.Println(fmt.Sprintf("%s: fetch message adr", p.Id, address))
	//	p.SendPacket(dat, address)
	}
	return nil
}

func (p *ProviderServer) SaveInPKI(path string) error {

	db, err := pki.OpenDatabase(path, "sqlite3")
	defer db.Close()

	if err != nil{
		return err
	}

	params := make(map[string]string)
	params["Id"] = "TEXT"
	params["Typ"] = "TEXT"
	params["Config"] = "BLOB"
	pki.CreateTable(db, "Providers", params)

	pubsBytes, err := config.MixPubsToBytes(p.Config)

	err = pki.InsertIntoTable(db, "Providers", p.Id, "Provider", pubsBytes)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProviderServer) Start() {
	defer p.Run()

	f, err := os.OpenFile("./logging/network_logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755)

	if err != nil{
		panic(err)
	}
	// defer f.Close()

	p.infoLogger = logging.NewInitLogger(f)
	p.errorLogger = logging.NewErrorLogger(f)
}

func (p *ProviderServer) Run() {

	defer p.listener.Close()
	finish := make(chan bool)

	go func() {
		p.infoLogger.Println(fmt.Sprintf("%s: Listening on %s", p.Id, p.Host + ":" + p.Port))
		p.ListenForIncomingConnections()
	}()

	<-finish
}

func NewProviderServer(id string, host string, port string, pubKey []byte, prvKey []byte, pkiPath string) (*ProviderServer, error) {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	providerServer := ProviderServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	providerServer.Config = config.MixPubs{Id: providerServer.Id, Host: providerServer.Host, Port: providerServer.Port, PubKey: providerServer.PubKey}

	err := providerServer.SaveInPKI(pkiPath)
	if err != nil {
		return nil, err
	}

	addr, err := helpers.ResolveTCPAddress(providerServer.Host, providerServer.Port)

	if err != nil {
		return nil, err
	}
	providerServer.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &providerServer, nil
}