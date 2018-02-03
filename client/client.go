/*
	Package client implements the class of a network client which can interact with a mix network.
*/

package client

import (
	"net"

	"anonymous-messaging/clientCore"
	"anonymous-messaging/networker"
	"anonymous-messaging/pki"
	"anonymous-messaging/config"
	"crypto/elliptic"

	log "github.com/sirupsen/logrus"
	"fmt"
	"anonymous-messaging/helpers"
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
	SendMessage(message string, recipient config.MixConfig)
	ProcessPacket(packet []byte)
	Start()
	ReadInMixnetPKI()
	ReadInClientsPKI()
}

type Client struct {
	Host string
	Port string
	clientCore.CryptoClient

	Listener *net.TCPListener

	PkiDir string
	// ActiveMixes  []config.MixConfig
	OtherClients []config.ClientConfig

	// Provider config.MixConfig
	Config config.ClientConfig

	token []byte

}

/*
	Start function creates the loggers for capturing the info and error logs;
	it reads the network and users information from the PKI database
	and starts the listening server. Function returns an error
	signaling whenever any operation was unsuccessful.
*/
func (c *Client) Start() error {


	err := c.ReadInClientsPKI(c.PkiDir)
	if err != nil{
		return err
	}

	err = c.ReadInMixnetPKI(c.PkiDir)
	if err != nil{
		return err
	}

	err = c.RegisterToProvider()
	if err != nil{
		return err
	}

	c.Run()
	return nil
}

/*
	SendMessage responsible for sending a real message. Takes as input the message string
	and the public information about the destination.
	The function generates a random path and a set of random values from exponential distribution.
	Given those values it triggers the encode function, which packs the message into the
	sphinx cryptographic packet format. Next, the encoded packet is combined with a
	flag signaling that this is a usual network packet, and passed to be send.
	The function returns an error if any issues occured.
*/
func (c *Client) SendMessage(message string, recipient config.ClientConfig) error {

	sphinxPacket, err := c.CreateSphinxPacket(message, recipient)
	if err != nil {
		return err
	}

	packetBytes, err := config.WrapWithFlag(COMM_FLAG, sphinxPacket)
	if err != nil{
		return err
	}

	err = c.Send(packetBytes, c.Provider.Host, c.Provider.Port)
	if err != nil {
		return err
	}
	return nil
}


/*
	Function opens a connection with selected network address
	and send the passed packet. If connection failed or
	the packet could not be send, an error is returned
*/
func (c *Client) Send(packet []byte, host string, port string) error {

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

/*
	Function responsible for running the listening process of the server;
	The clients listener accepts incoming connections and
	passes the incoming packets to the packet handler.
	If the connection could not be accepted an error
	is logged into the log files, but the function is not stopped
*/
func (c *Client) ListenForIncomingConnections() {
	for {
		conn, err := c.Listener.Accept()

		if err != nil {
			log.WithFields(log.Fields{"id" : c.Id}).Error(err)
		} else {
			go c.HandleConnection(conn)
		}
	}
}

/*
	Function handles the received packets; it checks the flag of the
	packet and schedules a corresponding process function;
	The potential errors are logged into the log files.
*/
func (c *Client) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	defer conn.Close()

	reqLen, err := conn.Read(buff)
	if err != nil {
		log.WithFields(log.Fields{"id" : c.Id}).Error(err)
		panic(err)
	}
	packet, err := config.GeneralPacketFromBytes(buff[:reqLen])
	if err != nil {
		log.WithFields(log.Fields{"id" : c.Id}).Error(err)
	}

	switch packet.Flag {
	case TOKEN_FLAG:
		c.RegisterToken(packet.Data)
		go func() {
			err = c.SendMessage("Hello world, this is me", c.OtherClients[0])
			if err != nil {
				log.WithFields(log.Fields{"id" : c.Id}).Error(err)
			}

			err = c.GetMessagesFromProvider()
			if err != nil {
				log.WithFields(log.Fields{"id" : c.Id}).Error(err)
			}
		}()
	case COMM_FLAG:
		_, err := c.ProcessPacket(packet.Data)
		if err != nil {
			log.WithFields(log.Fields{"id" : c.Id}).Error(err)
		}
		log.WithFields(log.Fields{"id" : c.Id}).Info(" Received new message")
	default:
		log.WithFields(log.Fields{"id" : c.Id}).Info(" Packet flag not recognised. Packet dropped.")
	}
}



/*
	Function stores the authentication token received from the provider
 */
func (c *Client) RegisterToken(token []byte) {
	c.token = token
	log.WithFields(log.Fields{"id" : c.Id}).Info(fmt.Sprintf(" Registered token %s", c.token))
}

/*
	ProcessPacket processes the received sphinx packet and returns the
	encapsulated message or error in case the processing
	was unsuccessful.
 */
func (c *Client) ProcessPacket(packet []byte) ([]byte, error) {
	log.WithFields(log.Fields{"id" : c.Id}).Info(" Processing packet")
	return packet, nil
}

/*
	RegisterToProvider allows the client to register with the selected provider.
	The client sends a special assignment packet, with its public information, to the provider
	or returns an error.
*/
func (c *Client) RegisterToProvider() error{

	log.WithFields(log.Fields{"id" : c.Id}).Info(" Sending request to provider to register")

	confBytes, err := config.ClientConfigToBytes(c.Config)
	if err != nil{
		return err
	}

	pktBytes, err := config.WrapWithFlag(ASSIGNE_FLAG, confBytes)
	if err != nil{
		return err
	}

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil{
		return err
	}
	return nil
}

/*
	GetMessagesFromProvider allows to fetch messages from the inbox stored by the
	provider. The client sends a pull packet to the provider, along with
	the authentication token. An error is returned if occurred.
*/
func (c *Client) GetMessagesFromProvider() error {
	pullRqs := config.PullRequest{Id: c.Id, Token: c.token}
	pullRqsBytes, err := config.PullRequestToBytes(pullRqs)
	if err != nil{
		return err
	}

	pktBytes, err := config.WrapWithFlag(PULL_FLAG, pullRqsBytes)
	if err != nil{
		return err
	}

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil{
		return err
	}

	return nil
}

/*
	Run opens the listener to start listening on clients host and port
 */
func (c *Client) Run() {
	defer c.Listener.Close()
	finish := make(chan bool)

	go func() {
		log.WithFields(log.Fields{"id" : c.Id}).Info(fmt.Sprintf("Listening on address %s", c.Host + ":" + c.Port))
		c.ListenForIncomingConnections()
	}()

	<-finish
}

/*
	ReadInMixnetPKI reads in the public information about active mixes
	from the PKI database and stores them locally. In case
	the connection or fetching data from the PKI went wrong,
	an error is returned.
*/
func (c *Client) ReadInMixnetPKI(pkiName string) error {
	log.WithFields(log.Fields{"id" : c.Id}).Info(fmt.Sprintf("Reading network information from the PKI: %s", pkiName))

	db, err := pki.OpenDatabase(pkiName, "sqlite3")

	if err != nil{
		return err
	}

	records, err := pki.QueryDatabase(db, "Pki", "Mix")

	if err != nil{
		return err
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			return err
		}

		pubs, err := config.MixConfigFromBytes(result["Config"].([]byte))
		if err != nil {
			return err
		}

		c.ActiveMixes = append(c.ActiveMixes, pubs)
	}

	log.WithFields(log.Fields{"id" : c.Id}).Info(" Network information uploaded")
	return nil
}

/*
	Function reads in the public information about users
	from the PKI database and stores them locally. In case
	the connection or fetching data from the PKI went wrong,
	an error is returned.
*/
func (c *Client) ReadInClientsPKI(pkiName string) error {
	log.WithFields(log.Fields{"id" : c.Id}).Info(fmt.Sprintf(" Reading network users information from the PKI: %s", pkiName))

	db, err := pki.OpenDatabase(pkiName, "sqlite3")

	if err != nil{
		return err
	}

	records, err := pki.QueryDatabase(db, "Pki", "Client")

	if err != nil {
		return err
	}

	for records.Next() {
		result := make(map[string]interface{})
		err := records.MapScan(result)

		if err != nil {
			return err
		}

		pubs, err := config.ClientConfigFromBytes(result["Config"].([]byte))
		if err != nil {
			return err
		}
		c.OtherClients = append(c.OtherClients, pubs)
	}
	log.WithFields(log.Fields{"id" : c.Id}).Info("  Information about other users uploaded")
	return nil
}


/*
	The constructor function to create an new client object.
	Function returns a new client object or an error, if occurred.
*/
func NewClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider config.MixConfig) (*Client, error) {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224(), Provider: provider}

	c := Client{Host: host, Port: port, CryptoClient: core, PkiDir: pkiDir}
	c.Config = config.ClientConfig{Id : c.Id, Host: c.Host, Port: c.Port, PubKey: c.PubKey, Provider: &c.Provider}

	configBytes, err := config.ClientConfigToBytes(c.Config)
	if err != nil{
		return nil, err
	}
	err = helpers.AddToDatabase(pkiDir, "Pki", c.Id, "Client", configBytes)
	if err != nil{
		return nil, err
	}


	addr, err := helpers.ResolveTCPAddress(c.Host, c.Port)
	if err != nil {
		return nil, err
	}

	c.Listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func NewTestClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider config.MixConfig) (*Client, error) {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224(), Provider: provider}
	c := Client{Host: host, Port: port, CryptoClient: core, PkiDir: pkiDir}
	c.Config = config.ClientConfig{Id : c.Id, Host: c.Host, Port: c.Port, PubKey: c.PubKey, Provider: &c.Provider}

	return &c, nil
}