/*
	Package client implements the class of a network client which can interact with a mix network.
*/

package client

import (
	"net"
	"os"

	"anonymous-messaging/clientCore"
	"anonymous-messaging/networker"
	"anonymous-messaging/pki"
	"anonymous-messaging/config"
	"github.com/jmoiron/sqlx"
	"crypto/elliptic"
	"anonymous-messaging/helpers"
	"anonymous-messaging/logging"
	"log"
	"fmt"
)

const (
	desiredRateParameter = 5
	pathLength           = 2
	ASSIGNE_FLAG = "\xA2"
	COMM_FLAG = "\xC6"
	TOKEN_FLAG = "xA9"
	PULL_FLAG = "\xFF"
)

type ClientIt interface {
	networker.NetworkClient
	networker.NetworkServer
	SendMessage(message string, recipient config.MixPubs)
	ProcessPacket(packet []byte)
	Start()
	ReadInMixnetPKI()
	ReadInClientsPKI()
	ConnectToPKI(dbName string) *sqlx.DB
}

type Client struct {
	Host string
	Port string
	clientCore.CryptoClient

	listener *net.TCPListener

	pkiDir string
	ActiveMixes  []config.MixPubs
	OtherClients []config.ClientPubs

	Provider config.MixPubs
	Config config.ClientPubs

	infoLogger *log.Logger
	errorLogger *log.Logger

	token []byte

}

func (c *Client) SendMessage(message string, recipient config.ClientPubs) {


	path := c.buildPath(recipient)
	delays := c.GenerateDelaySequence(desiredRateParameter, path.Len())

	sphinxPacket, err := c.EncodeMessage(message, path, delays)
	if err != nil{
		panic(err)
	}

	packet := config.GeneralPacket{Flag: COMM_FLAG, Data: sphinxPacket}
	packetBytes, err := config.GeneralPacketToBytes(packet)
	if err != nil{
		c.errorLogger.Println(err)
		panic(err)
	}

	err = c.Send(packetBytes, path.IngressProvider.Host, path.IngressProvider.Port)
	if err != nil{
		c.errorLogger.Println(err)
		panic(err)
	}
}

func (c *Client) buildPath(recipient config.ClientPubs) config.E2EPath {
	var path config.E2EPath

	mixSeq := c.GetRandomMixSequence(c.ActiveMixes, pathLength)
	path.IngressProvider = c.Provider
	path.Mixes = mixSeq
	path.EgressProvider = *recipient.Provider
	path.Recipient = recipient

	return path
}

func (c *Client) Send(packet []byte, host string, port string) error {

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		c.errorLogger.Println(err)
		os.Exit(1)
	} else {
		defer conn.Close()
	}

	_, err = conn.Write(packet)
	return err
}

func (c *Client) ListenForIncomingConnections() {
	for {
		conn, err := c.listener.Accept()

		if err != nil {
			c.errorLogger.Println(err)
			os.Exit(1)
		}
		go c.HandleConnection(conn)
	}
}

func (c *Client) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)

	reqLen, err := conn.Read(buff)
	if err != nil {
		c.errorLogger.Println(err)
		panic(err)
	}
	packet, err := config.GeneralPacketFromBytes(buff[:reqLen])
	if err != nil {
		c.errorLogger.Println(err)
	}
	switch packet.Flag {
	case TOKEN_FLAG:
		c.RegisterToken(packet.Data)
		go func() {
			c.SendMessage("Hello world, this is me", c.OtherClients[0])
			err = c.GetMessagesFromProvider()
			if err != nil {
				c.errorLogger.Println(err)
			}
		}()
	case COMM_FLAG:
		c.ProcessPacket(packet.Data)
	default:
		c.infoLogger.Println(fmt.Sprintf("%s: Packet flag not recognised. Packet dropped.", c.Id))
	}

	conn.Close()
}

func (c *Client) RegisterToken(token []byte) {
	c.token = token
	c.infoLogger.Println(fmt.Sprintf("%s: Registered token %s", c.Id, c.token))
}

func (c *Client) ProcessPacket(packet []byte) []byte {
	c.infoLogger.Println("%s: Processing packet: %s", c.Id, packet )
	return packet
}

func (c *Client) Start() {

	f, err := os.OpenFile("./logging/client_logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755)

	if err != nil{
		panic(err)
	}

	c.infoLogger = logging.NewInitLogger(f)
	c.errorLogger = logging.NewErrorLogger(f)

	defer c.Run()

	c.ReadInClientsPKI(c.pkiDir)
	c.ReadInMixnetPKI(c.pkiDir)
	c.RegisterToProvider()

}

func (c *Client) RegisterToProvider() error{

	c.infoLogger.Println(fmt.Sprintf("%s: Sending to provider", c.Id))

	confBytes, err := config.ClientPubsToBytes(c.Config)
	if err != nil{
		return err
	}

	pkt := config.GeneralPacket{Flag: ASSIGNE_FLAG, Data: confBytes}
	pktBytes, err := config.GeneralPacketToBytes(pkt)
	if err != nil{
		return err
	}

	c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	return nil
}

func (c *Client) GetMessagesFromProvider() error {
	pullRqs := config.PullRequest{Id: c.Id, Token: c.token}
	pullRqsBytes, err := config.PullRequestToBytes(pullRqs)
	if err != nil{
		return err
	}

	pkt := config.GeneralPacket{Flag: PULL_FLAG, Data: pullRqsBytes}
	pktBytes, err := config.GeneralPacketToBytes(pkt)
	if err != nil{
		return err
	}

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil{
		return err
	}

	return nil
}

func (c *Client) Run() {
	defer c.listener.Close()
	finish := make(chan bool)

	go func() {
		c.infoLogger.Println(fmt.Sprintf("%s: Listening on address %s", c.Id, c.Host + ":" + c.Port))
		c.ListenForIncomingConnections()
	}()

	<-finish
}

func (c *Client) ReadInMixnetPKI(pkiName string) {
	c.infoLogger.Println(fmt.Sprintf("%s: Reading network information from the PKI: %s", c.Id, pkiName))

	db, err := c.ConnectToPKI(pkiName)

	if err != nil{
		c.errorLogger.Println(err)
		panic(err)
	}

	records, err := pki.QueryDatabase(db, "Mixes")

	if err != nil{
		c.errorLogger.Println(err)
		panic(err)
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			c.errorLogger.Println(err)
			panic(err)

		}
		pubs, err := config.MixPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			c.errorLogger.Println(err)
			panic(err)
		}

		c.ActiveMixes = append(c.ActiveMixes, pubs)
	}
	c.infoLogger.Println(fmt.Sprintf("%s: Network information uploaded", c.Id))

}

func (c *Client) ReadInClientsPKI(pkiName string) {
	c.infoLogger.Println(fmt.Sprintf("%s: Reading network users information from the PKI: %s", c.Id, pkiName))

	db, err := c.ConnectToPKI(pkiName)

	if err != nil{
		c.errorLogger.Println(err)
		panic(err)
	}

	records, err := pki.QueryDatabase(db, "Clients")

	if err != nil {
		c.errorLogger.Println(err)
		panic(err)
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			c.errorLogger.Println(err)
			panic(err)
		}

		pubs, err := config.ClientPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			c.errorLogger.Println(err)
			panic(err)
		}
		c.OtherClients = append(c.OtherClients, pubs)
	}
	c.infoLogger.Println(fmt.Sprintf("%s: Network users information uploaded", c.Id))
}

func (c *Client) ConnectToPKI(dbName string) (*sqlx.DB, error) {
	return pki.OpenDatabase(dbName, "sqlite3")
}


func NewClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider config.MixPubs) (*Client, error) {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224()}

	c := Client{Host: host, Port: port, CryptoClient: core, Provider: provider, pkiDir: pkiDir}
	c.Config = config.ClientPubs{Id : c.Id, Host: c.Host, Port: c.Port, PubKey: c.PubKey, Provider: &c.Provider}

	configBytes, err := config.ClientPubsToBytes(c.Config)
	if err != nil{
		return nil, err
	}
	err = helpers.AddToDatabase(pkiDir, "Clients", c.Id, "Client", configBytes)
	if err != nil{
		return nil, err
	}


	addr, err := helpers.ResolveTCPAddress(c.Host, c.Port)
	if err != nil {
		return nil, err
	}
	c.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		return nil, err
	}
	return &c, nil
}
