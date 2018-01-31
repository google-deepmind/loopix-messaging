package server

import (
	"anonymous-messaging/node"
	"net"
	"anonymous-messaging/networker"
	"os"
	"fmt"
	"bytes"
	"anonymous-messaging/config"
	"io/ioutil"
	"anonymous-messaging/helpers"
	"anonymous-messaging/logging"
	"log"
	"errors"
	"anonymous-messaging/sphinx"
)

const (
	ASSIGNE_FLAG = "\xA2"
	COMM_FLAG = "\xC6"
	TOKEN_FLAG = "xA9"
	PULL_FLAG = "\xFF"
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

// Start function creates the loggers for capturing the info and error logs
// and starts the listening server. Function returns an error
// signaling whether any operation was unsuccessful
func (p *ProviderServer) Start() error{
	defer p.Run()

	f, err := os.OpenFile("./logging/network_logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755)
	if err != nil{
		return err
	}

	p.infoLogger = logging.NewInitLogger(f)
	p.errorLogger = logging.NewErrorLogger(f)
	return nil
}

// Function opens the listener to start listening on provider's host and port
func (p *ProviderServer) Run() {

	defer p.listener.Close()
	finish := make(chan bool)

	go func() {
		p.infoLogger.Println(fmt.Sprintf("%s: Listening on %s", p.Id, p.Host + ":" + p.Port))
		p.ListenForIncomingConnections()
	}()

	<-finish
}

// Function processes the received sphinx packet, performs the
// unwrapping operation and checks whether the packet should be
// forwarded or stored. If the processing was unsuccessful and error is returned.
func (p *ProviderServer) ReceivedPacket(packet []byte) error{
	p.infoLogger.Println(fmt.Sprintf("%s: Received new packet", p.Id))

	c := make(chan []byte)
	cAdr := make(chan sphinx.Hop)
	cFlag := make(chan string) // CHANGE BACK TO HOP, BECAUSE YOU NEED IT
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
		err = p.ForwardPacket(dePacket, nextHop.Address)
		if err != nil{
			return err
		}
	case "\xF0":
		err = p.StoreMessage(dePacket, nextHop.Id, "TMP_MESSAGE_ID")
		if err != nil{
			return err
		}
	default:
		p.infoLogger.Println(fmt.Sprintf("%s: Sphinx packet flag not recognised", p.Id))
	}

	return nil
}

func (p *ProviderServer) ForwardPacket(sphinxPacket []byte, address string) error{
	packet := config.GeneralPacket{Flag:COMM_FLAG, Data: sphinxPacket}
	packetBytes, err := config.GeneralPacketToBytes(packet)

	err = p.Send(packetBytes, address)
	if err != nil{
		return err
	}
	p.infoLogger.Println(fmt.Sprintf("%s: Forwarded packet", p.Id))
	return nil
}

// Function opens a connection with selected network address
// and send the passed packet. If connection failed or
// the packet could not be send, an error is returned
func (p *ProviderServer) Send(packet []byte, address string) error {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Write(packet)
	return nil
}


// Function responsible for running the listening process of the server;
// The providers listener accepts incoming connections and
// passes the incoming packets to the packet handler.
// If the connection could not be accepted an error
// is logged into the log files, but the function is not stopped
func (p *ProviderServer) ListenForIncomingConnections() {
	for {
		conn, err := p.listener.Accept()

		if err != nil {
			p.errorLogger.Println(err)
		} else {
			p.infoLogger.Println(fmt.Sprintf("%s: Received new connection from %s", p.Id, conn.RemoteAddr()))
			go p.HandleConnection(conn)
		}
	}
}

// Function handles the received packets; it checks the flag of the
// packet and schedules a corresponding process function;
// The potential errors are logged into the log files.
func (p *ProviderServer) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		p.errorLogger.Println(err)
	}

	packet, err := config.GeneralPacketFromBytes(buff[:reqLen])
	if err != nil {
		p.errorLogger.Println(err)
	}

	switch packet.Flag {
	case ASSIGNE_FLAG:
		err = p.HandleAssignRequest(packet.Data)
		if err != nil {
			p.errorLogger.Println(err)
		}
	case COMM_FLAG:
		err = p.ReceivedPacket(packet.Data)
		if err != nil {
			p.errorLogger.Println(err)
		}
	case PULL_FLAG:
		err = p.HandlePullRequest(packet.Data)
		if err != nil{
			p.errorLogger.Println(err)
		}
	default:
		p.infoLogger.Println(fmt.Sprintf("%s : Packet flag not recognised. Packet dropped.", p.Id))
	}

	conn.Close()
}

// Function generates a fresh authentication token and saves it together with client's public configuration data
// in the list of all registered clients. After the client is registered the function creates an inbox directory
// for the client's inbox, in which clients messages will be stored.
func (p *ProviderServer) RegisterNewClient(clientBytes []byte) ([]byte, string, error){
	clientConf, err := config.ClientPubsFromBytes(clientBytes)
	if err != nil{
		return nil, "", err
	}

	token := []byte("TMP_Token" + clientConf.Id)
	record := ClientRecord{Id: clientConf.Id, Host: clientConf.Host, Port: clientConf.Port, PubKey: clientConf.PubKey, Token: token}
	p.assignedClients[clientConf.Id] = record
	address := clientConf.Host + ":" + clientConf.Port

	path := fmt.Sprintf("./inboxes/%s", clientConf.Id)
	exists, err := helpers.DirExists(path)
	if err != nil{
		return nil, "", err
	}
	if exists == false {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, "", err
		}
	}

	return token, address, nil
}


// Function is responsible for handling the registration request from the client.
// it registers the client in the list of all registered clients and send
// an authentication token back to the client.
func (p *ProviderServer) HandleAssignRequest(packet []byte) error {
	p.infoLogger.Println(fmt.Sprintf("%s : Received assign request from the client.", p.Id))

	token, adr, err := p.RegisterNewClient(packet)
	if err != nil {
		return err
	}
	tokenPacket := config.GeneralPacket{Flag: TOKEN_FLAG, Data: token}
	tokenBytes, err := config.GeneralPacketToBytes(tokenPacket)
	if err != nil {
		return err
	}

	err = p.Send(tokenBytes, adr)
	if err != nil {
		return err
	}

	return nil
}

// Function is responsible for handling the pull request received from the client.
// It first authenticates the client, by checking if the received token is valid.
// If yes, the function triggers the function for checking client's inbox
// and sending buffered messages. Otherwise, an error is returned.
func (p *ProviderServer) HandlePullRequest(rqsBytes []byte) error {
	request, err := config.PullRequestFromBytes(rqsBytes)
	if err != nil {
		return err
	}

	p.infoLogger.Println(fmt.Sprintf("%s : Handling a pull request.", p.Id))
	p.infoLogger.Println(fmt.Sprintf("%s : Request: %s %s", p.Id, request.Id, string(request.Token)))

	if p.AuthenticateUser(request.Id, request.Token) == true{
		err = p.FetchMessages(request.Id)
		if err != nil {
			return err
		}
	} else {
		p.infoLogger.Println(fmt.Sprintf("%s: Authentication went wrong", p.Id))
		return errors.New("Authentication went wrong! ")
	}
	return nil
}

// Function compares the authentication token received from the client with
// the one stored by the provider. If tokens are the same, it returns true
// and false otherwise.
func (p *ProviderServer) AuthenticateUser(clientId string, clientToken []byte) bool{

	if bytes.Compare(p.assignedClients[clientId].Token, clientToken) == 0 {
		return true
	}
	return false
}

// Function fetches messages from the requested inbox.
// If inbox contains any stored messages, all of them
// are send to the client one by one. Function
// returns an error in case any operation went wrong.
func (p *ProviderServer) FetchMessages(clientId string) error{


	path := fmt.Sprintf("./inboxes/%s", clientId)
	_, err := os.Stat(path)
	if err != nil{
		return err
	}

	files, err := ioutil.ReadDir(path)
	if err != nil{
		return err
	}

	for _, f := range files {
		dat, err := ioutil.ReadFile(path + "/" + f.Name())
		if err !=nil {
			return err
		}

		address := p.assignedClients[clientId].Host + ":" + p.assignedClients[clientId].Port
		p.infoLogger.Println(fmt.Sprintf("%s: Found stored message for address %s", p.Id, address))
		msg := config.GeneralPacket{Flag: COMM_FLAG, Data: dat}
		msgBytes, err := config.GeneralPacketToBytes(msg)
		if err !=nil {
			return err
		}
		err = p.Send(msgBytes, address)
		if err !=nil {
			return err
		}

	}
	p.infoLogger.Println(fmt.Sprintf("%s: All messages for address fetched", p.Id))
	return nil
}

// Function saves the given message in the inbox defined by the given id.
// If the inbox address does not exist or writing into the inbox was unsuccessful
// the function returns an error
func (p *ProviderServer) StoreMessage(message []byte, inboxId string, messageId string) error {
	path := fmt.Sprintf("./inboxes/%s", inboxId)
	fileName := path + "/" + messageId

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(message)
	if err != nil {
		return err
	}

	p.infoLogger.Println(fmt.Sprintf("%s: stored message for %s", p.Id, inboxId))
	return nil
}

// The constructor function to create an new provider object.
// Function returns a new provider object or an error, if occurred.
func NewProviderServer(id string, host string, port string, pubKey []byte, prvKey []byte, pkiPath string) (*ProviderServer, error) {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	providerServer := ProviderServer{Id: id, Host: host, Port: port, Mix: node, listener: nil}
	providerServer.Config = config.MixPubs{Id: providerServer.Id, Host: providerServer.Host, Port: providerServer.Port, PubKey: providerServer.PubKey}
	providerServer.assignedClients = make(map[string]ClientRecord)

	configBytes, err := config.MixPubsToBytes(providerServer.Config)
	if err != nil{
		return nil, err
	}
	err = helpers.AddToDatabase(pkiPath, "Providers", providerServer.Id, "Provider", configBytes)
	if err != nil{
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