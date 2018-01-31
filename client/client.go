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

func (c *Client) SendMessage(message string, recipient config.ClientPubs) error {


	path := c.buildPath(recipient)
	delays := c.GenerateDelaySequence(desiredRateParameter, path.Len())

	sphinxPacket, err := c.EncodeMessage(message, path, delays)
	if err != nil{
		return err
	}

	packet := config.GeneralPacket{Flag: COMM_FLAG, Data: sphinxPacket}
	packetBytes, err := config.GeneralPacketToBytes(packet)
	if err != nil{
		return err
	}

	err = c.Send(packetBytes, path.IngressProvider.Host, path.IngressProvider.Port)
	if err != nil {
		return err
	}

	return nil
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
			err = c.SendMessage("Hello world, this is me", c.OtherClients[0])
			if err != nil {
				c.errorLogger.Println(err)
			}

			err = c.GetMessagesFromProvider()
			if err != nil {
				c.errorLogger.Println(err)
			}
		}()
	case COMM_FLAG:
		newPkt, err := c.ProcessPacket(packet.Data)
		if err != nil {
			c.errorLogger.Println(err)
		}
		c.infoLogger.Println(fmt.Sprintf("%s: Received message: %s", c.Id, newPkt))
	default:
		c.infoLogger.Println(fmt.Sprintf("%s: Packet flag not recognised. Packet dropped.", c.Id))
	}

	conn.Close()
}

func (c *Client) RegisterToken(token []byte) {
	c.token = token
	c.infoLogger.Println(fmt.Sprintf("%s: Registered token %s", c.Id, c.token))
}

func (c *Client) ProcessPacket(packet []byte) ([]byte, error) {
	c.infoLogger.Println("%s: Processing packet: %s", c.Id, packet )
	return packet, nil
}

func (c *Client) Start() error {

	f, err := os.OpenFile("./logging/client_logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755)

	if err != nil{
		return err
	}

	c.infoLogger = logging.NewInitLogger(f)
	c.errorLogger = logging.NewErrorLogger(f)

	defer c.Run()

	err = c.ReadInClientsPKI(c.pkiDir)
	if err != nil{
		return err
	}

	err = c.ReadInMixnetPKI(c.pkiDir)
	if err != nil{
		return err
	}

	err = c.RegisterToProvider()
	if err != nil{
		return err
	}
	return nil
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

func (c *Client) ReadInMixnetPKI(pkiName string) error {
	c.infoLogger.Println(fmt.Sprintf("%s: Reading network information from the PKI: %s", c.Id, pkiName))

	db, err := pki.OpenDatabase(pkiName, "sqlite3")

	if err != nil{
		return err
	}

	records, err := pki.QueryDatabase(db, "Mixes")

	if err != nil{
		return err
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			return err
		}

		pubs, err := config.MixPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			return err
		}

		c.ActiveMixes = append(c.ActiveMixes, pubs)
	}
	c.infoLogger.Println(fmt.Sprintf("%s: Network information uploaded", c.Id))
	return nil
}

func (c *Client) ReadInClientsPKI(pkiName string) error {
	c.infoLogger.Println(fmt.Sprintf("%s: Reading network users information from the PKI: %s", c.Id, pkiName))

	db, err := pki.OpenDatabase(pkiName, "sqlite3")

	if err != nil{
		return err
	}

	records, err := pki.QueryDatabase(db, "Clients")

	if err != nil {
		return err
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			return err
		}

		pubs, err := config.ClientPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			return err
		}
		c.OtherClients = append(c.OtherClients, pubs)
	}
	c.infoLogger.Println(fmt.Sprintf("%s: Network users information uploaded", c.Id))
	return nil
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
