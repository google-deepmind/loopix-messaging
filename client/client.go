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


// Function responsible for sending a real message. Takes as input the message string
// and the public information about the destination.
// The function generates a random path and a set of random values from exponential distribution.
// Given those values it triggeres the encode function, which packs the message into the
// sphinx cryptographic packet format. Next, the encoded packet is combined with a
// flag signaling that this is a usual network packet, and passed to be send.
// The function returns an error if any issues occured.
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

// Function build a path containing the sender's provider,
// a sequence (of length pre-defined in a config file) of randomly
// selected mixes and the recipient's provider
func (c *Client) buildPath(recipient config.ClientPubs) config.E2EPath {
	var path config.E2EPath

	mixSeq := c.GetRandomMixSequence(c.ActiveMixes, pathLength)
	path.IngressProvider = c.Provider
	path.Mixes = mixSeq
	path.EgressProvider = *recipient.Provider
	path.Recipient = recipient

	return path
}

// Function opens a connection with selected network address
// and send the passed packet. If connection failed or
// the packet could not be send, an error is returned
func (c *Client) Send(packet []byte, host string, port string) error {

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return err
	} else {
		defer conn.Close()
	}

	_, err = conn.Write(packet)
	return err
}

// Function responsible for running the listening process of the server;
// The clients listener accepts incoming connections and
// passes the incoming packets to the packet handler.
// If the connection could not be accepted an error
// is logged into the log files, but the function is not stopped
func (c *Client) ListenForIncomingConnections() {
	for {
		conn, err := c.listener.Accept()

		if err != nil {
			c.errorLogger.Println(err)
		} else {
			go c.HandleConnection(conn)
		}
	}
}

// Function handles the received packets; it checks the flag of the
// packet and schedules a corresponding process function;
// The potential errors are logged into the log files.
func (c *Client) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	defer conn.Close()

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
}


// Function stores the authentication token received from the provider
func (c *Client) RegisterToken(token []byte) {
	c.token = token
	c.infoLogger.Println(fmt.Sprintf("%s: Registered token %s", c.Id, c.token))
}

// Function processes the sphinx packet and returns the
// encapsulated message or error in case the processing
// was unsuccessful
func (c *Client) ProcessPacket(packet []byte) ([]byte, error) {
	c.infoLogger.Println("%s: Processing packet: %s", c.Id, packet )
	return packet, nil
}

// Start function creates the loggers for capturing the info and error logs;
// it reads the network and users information from the PKI database
// and starts the listening server. Function returns an error
// signaling whenther any operation was unsuccessful
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

// Function allows the client to register with the selected provider.
// The client sends a special assignment packet, with its public information, to the provider
// or returns an error
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

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil{
		return err
	}
	return nil
}

// Function allows to fetch messages from the inbox stored by the
// provider. The client sends a pull packet to the provider, along with
// the authentication token. An error is returned if occurred.
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

// Function opens the listener to start listening on clients host and port
func (c *Client) Run() {
	defer c.listener.Close()
	finish := make(chan bool)

	go func() {
		c.infoLogger.Println(fmt.Sprintf("%s: Listening on address %s", c.Id, c.Host + ":" + c.Port))
		c.ListenForIncomingConnections()
	}()

	<-finish
}

// Function reads in the public information about active mixes
// from the PKI database and stores them locally. In case
// the connection or fetching data from the PKI went wrong,
// an error is returned.
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

// Function reads in the public information about users
// from the PKI database and stores them locally. In case
// the connection or fetching data from the PKI went wrong,
// an error is returned.
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


// The constructor function to create an new client object.
// Function returns a new client object or an error, if occured. 
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
