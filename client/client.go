/*
	Package client implements the class of a network client which can interact with a mix network.
*/

package client

import (
	"anonymous-messaging/clientCore"
	"anonymous-messaging/config"
	"anonymous-messaging/helpers"
	"anonymous-messaging/logging"
	"anonymous-messaging/networker"
	"anonymous-messaging/pki"

	"github.com/protobuf/proto"

	"crypto/elliptic"
	"math"
	"net"
	"time"
)

var logLocal = logging.PackageLogger()

const (
	// the parameter of the exponential distribution which defines the rate of sending by client
	// the desiredRateParameter is the reciprocal of the expected value of the exponential distribution
	desiredRateParameter = 0.2
	// the rate at which clients are querying the provider for received packets. fetchRate value is the
	// parameter of an exponential distribution, and is the reciprocal of the expected value of the exp. distribution
	fetchRate  = 0.01
	assignFlag = "\xA2"
	commFlag   = "\xc6"
	tokenFlag  = "xa9"
	pullFlag   = "\xff"
)

type ClientIt interface {
	networker.NetworkClient
	networker.NetworkServer
	SendMessage(message string, recipient config.MixConfig)
	ProcessPacket(packet []byte)
	Start()
	ReadInNetworkFromPKI()
	ReadInClientsPKI()
}

type Client struct {
	Host string
	Port string

	clientCore.CryptoClient
	Listener *net.TCPListener

	PkiDir string

	Config config.ClientConfig
	token  []byte

	OutQueue         chan []byte
	registrationDone chan bool
}

// Start function creates the loggers for capturing the info and error logs;
// it reads the network and users information from the PKI database
// and starts the listening server. Function returns an error
// signaling whenever any operation was unsuccessful.
func (c *Client) Start() error {

	c.OutQueue = make(chan []byte)
	c.registrationDone = make(chan bool)

	err := c.ReadInNetworkFromPKI(c.PkiDir)
	if err != nil {
		logLocal.WithError(err).Error("Error during reading in network PKI")
		return err
	}

	go func() {
		for {
			select {
			case <-c.registrationDone:
				return
			default:
				err = c.SendRegisterMessageToProvider()
				if err != nil {
					logLocal.WithError(err).Error("Error during registration to provider", err)
				}
				time.Sleep(60 * time.Second)
			}
		}
	}()

	c.Run()
	return nil
}

// SendMessage responsible for sending a real message. Takes as input the message string
// and the public information about the destination.
// The function generates a random path and a set of random values from exponential distribution.
// Given those values it triggers the encode function, which packs the message into the
// sphinx cryptographic packet format. Next, the encoded packet is combined with a
// flag signaling that this is a usual network packet, and passed to be send.
// The function returns an error if any issues occurred.
func (c *Client) SendMessage(message string, recipient config.ClientConfig) error {

	sphinxPacket, err := c.CreateSphinxPacket(message, recipient)
	if err != nil {
		logLocal.WithError(err).Error("Error in sending message - create sphinx packet returned an error")
		return err
	}

	packetBytes, err := config.WrapWithFlag(commFlag, sphinxPacket)
	if err != nil {
		logLocal.WithError(err).Error("Error in sending message - wrap with flag returned an error")
		return err
	}
	c.OutQueue <- packetBytes
	return nil
}

// Send opens a connection with selected network address
// and send the passed packet. If connection failed or
// the packet could not be send, an error is returned
func (c *Client) Send(packet []byte, host string, port string) error {

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		logLocal.WithError(err).Error("Error in send - dial returned an error")
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

// ListenForIncomingConnections responsible for running the listening process of the server;
// The clients listener accepts incoming connections and
// passes the incoming packets to the packet handler.
// If the connection could not be accepted an error
// is logged into the log files, but the function is not stopped
func (c *Client) ListenForIncomingConnections() {
	for {
		conn, err := c.Listener.Accept()

		if err != nil {
			logLocal.WithError(err).Error(err)
		} else {
			go c.HandleConnection(conn)
		}
	}
}

// HandleConnection handles the received packets; it checks the flag of the
// packet and schedules a corresponding process function;
// The potential errors are logged into the log files.
func (c *Client) HandleConnection(conn net.Conn) {

	buff := make([]byte, 1024)
	defer conn.Close()

	reqLen, err := conn.Read(buff)
	if err != nil {
		logLocal.WithError(err).Error("Error while reading incoming connection")
		panic(err)
	}
	var packet config.GeneralPacket
	err = proto.Unmarshal(buff[:reqLen], &packet)
	if err != nil {
		logLocal.WithError(err).Error("Error in unmarshal incoming packet")
	}

	switch packet.Flag {
	case tokenFlag:
		c.RegisterToken(packet.Data)
		go func() {
			err := c.ControlOutQueue()
			if err != nil {
				logLocal.WithError(err).Panic("Error in the controller of the outgoing packets queue. Possible security threat.")
			}
		}()

		go func() {
			c.ControlMessagingFetching()
		}()
	case commFlag:
		_, err := c.ProcessPacket(packet.Data)
		if err != nil {
			logLocal.WithError(err).Error("Error in processing received packet")
		}
		logLocal.Info("Received new message")
	default:
		logLocal.Info("Packet flag not recognised. Packet dropped.")
	}
}

// RegisterToken stores the authentication token received from the provider
func (c *Client) RegisterToken(token []byte) {
	c.token = token
	logLocal.Infof(" Registered token %s", c.token)
	c.registrationDone <- true
}

// ProcessPacket processes the received sphinx packet and returns the
// encapsulated message or error in case the processing
// was unsuccessful.
func (c *Client) ProcessPacket(packet []byte) ([]byte, error) {
	logLocal.Info(" Processing packet")
	return packet, nil
}

// SendRegisterMessageToProvider allows the client to register with the selected provider.
// The client sends a special assignment packet, with its public information, to the provider
// or returns an error.
func (c *Client) SendRegisterMessageToProvider() error {

	logLocal.Info("Sending request to provider to register")

	confBytes, err := proto.Marshal(&c.Config)
	if err != nil {
		logLocal.WithError(err).Error("Error in register provider - marshal of provider config returned an error")
		return err
	}

	pktBytes, err := config.WrapWithFlag(assignFlag, confBytes)
	if err != nil {
		logLocal.WithError(err).Error("Error in register provider - wrap with flag returned an error")
		return err
	}

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil {
		logLocal.WithError(err).Error("Error in register provider - send registration packet returned an error")
		return err
	}
	return nil
}

// GetMessagesFromProvider allows to fetch messages from the inbox stored by the
// provider. The client sends a pull packet to the provider, along with
// the authentication token. An error is returned if occurred.
func (c *Client) GetMessagesFromProvider() error {
	pullRqs := config.PullRequest{ClientId: c.Id, Token: c.token}
	pullRqsBytes, err := proto.Marshal(&pullRqs)
	if err != nil {
		logLocal.WithError(err).Error("Error in register provider - marshal of pull request returned an error")
		return err
	}

	pktBytes, err := config.WrapWithFlag(pullFlag, pullRqsBytes)
	if err != nil {
		logLocal.WithError(err).Error("Error in register provider - marshal of provider config returned an error")
		return err
	}

	err = c.Send(pktBytes, c.Provider.Host, c.Provider.Port)
	if err != nil {
		return err
	}

	return nil
}

// Run opens the listener to start listening on clients host and port
func (c *Client) Run() {
	defer c.Listener.Close()
	finish := make(chan bool)

	go func() {
		c.FakeAdding()
	}()

	go func() {
		logLocal.Infof("Listening on address %s", c.Host+":"+c.Port)
		c.ListenForIncomingConnections()
	}()

	<-finish
}

func (c *Client) FakeAdding() {
	logLocal.Info("Started fake adding")
	for {
		sphinxPacket, err := c.CreateSphinxPacket("hello world", c.Config)
		if err != nil {
			logLocal.Info("Something went wrong")
		}
		packet, err := config.WrapWithFlag(commFlag, sphinxPacket)
		if err != nil {
			logLocal.Info("Something went wrong")
		}
		c.OutQueue <- packet
		logLocal.Info("Added packet")
		time.Sleep(10 * time.Second)
	}
}

func (c *Client) ControlOutQueue() error {
	logLocal.Info("Queue controller started")
	for {
		select {
		case realPacket := <-c.OutQueue:
			c.Send(realPacket, c.Provider.Host, c.Provider.Port)
			logLocal.Info("Real packet was sent")
		default:
			dummyPacket, err := c.CreateCoverMessage()
			if err != nil {
				return err
			}
			c.Send(dummyPacket, c.Provider.Host, c.Provider.Port)
			logLocal.Info("OutQueue empty. Dummy packet sent.")
		}
		delaySec, err := helpers.RandomExponential(desiredRateParameter)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(int64(delaySec*math.Pow10(9))) * time.Nanosecond)
	}
	return nil
}

func (c *Client) ControlMessagingFetching() {
	for {
		c.GetMessagesFromProvider()
		logLocal.Info("Sent request to provider to fetch messages")

		timeout, err := helpers.RandomExponential(fetchRate)
		if err != nil {
			logLocal.Error("Error in ControlMessagingFetching - generating random exp. value failed")
		}
		time.Sleep(time.Duration(int64(timeout*math.Pow10(9))) * time.Nanosecond)
	}
}

// CreateCoverMessage packs a dummy message into a Sphinx packet.
// The dummy message is a loop message.
// TODO: change to a drop cover message instead of a loop.
func (c *Client) CreateCoverMessage() ([]byte, error) {
	dummyLoad := "DummyPayloadMessage"
	sphinxPacket, err := c.CreateSphinxPacket(dummyLoad, c.Config)
	if err != nil {
		return nil, err
	}

	packetBytes, err := config.WrapWithFlag(commFlag, sphinxPacket)
	if err != nil {
		return nil, err
	}
	return packetBytes, nil
}

// ReadInNetworkFromPKI reads in the public information about active mixes
// from the PKI database and stores them locally. In case
// the connection or fetching data from the PKI went wrong,
// an error is returned.
func (c *Client) ReadInNetworkFromPKI(pkiName string) error {
	logLocal.Infof("Reading network information from the PKI: %s", pkiName)

	db, err := pki.OpenDatabase(pkiName, "sqlite3")

	if err != nil {
		return err
	}

	recordsMixes, err := pki.QueryDatabase(db, "Pki", "Mix")
	if err != nil {
		logLocal.WithError(err).Error("Error during querying the Mixes PKI")
		return err
	}

	for recordsMixes.Next() {
		result := make(map[string]interface{})
		err := recordsMixes.MapScan(result)
		if err != nil {
			logLocal.WithError(err).Error("Error during mixes record mapping PKI")
			return err
		}

		var mixConfig config.MixConfig
		err = proto.Unmarshal(result["Config"].([]byte), &mixConfig)
		if err != nil {
			logLocal.WithError(err).Error("Error during unmarshal function for mix config")
			return err
		}
		c.Network.Mixes = append(c.Network.Mixes, mixConfig)
	}

	recordsProviders, err := pki.QueryDatabase(db, "Pki", "Provider")
	if err != nil {
		logLocal.WithError(err).Error("Error during querying the Providers PKI")
		return err
	}
	for recordsProviders.Next() {
		result := make(map[string]interface{})
		err := recordsProviders.MapScan(result)

		if err != nil {
			logLocal.WithError(err).Error("Error during providers record mapping PKI")
			return err
		}

		var prvConfig config.MixConfig
		err = proto.Unmarshal(result["Config"].([]byte), &prvConfig)
		if err != nil {
			logLocal.WithError(err).Error("Error during unmarshal function for provider config")
			return err
		}

		c.Network.Providers = append(c.Network.Providers, prvConfig)
	}
	logLocal.Info("Network information uploaded")

	return nil
}

// The constructor function to create an new client object.
// Function returns a new client object or an error, if occurred.
func NewClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider config.MixConfig) (*Client, error) {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224(), Provider: provider}

	c := Client{Host: host, Port: port, CryptoClient: core, PkiDir: pkiDir}
	c.Config = config.ClientConfig{Id: c.Id, Host: c.Host, Port: c.Port, PubKey: c.PubKey, Provider: &c.Provider}

	configBytes, err := proto.Marshal(&c.Config)

	if err != nil {
		return nil, err
	}
	err = helpers.AddToDatabase(pkiDir, "Pki", c.Id, "Client", configBytes)
	if err != nil {
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

// NewTestClient constructs a client object, which can be used for testing. The object contains the crypto core
// and the top-level of client, but does not involve networking and starting a listener.
func NewTestClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider config.MixConfig) (*Client, error) {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224(), Provider: provider}
	c := Client{Host: host, Port: port, CryptoClient: core, PkiDir: pkiDir}
	c.Config = config.ClientConfig{Id: c.Id, Host: c.Host, Port: c.Port, PubKey: c.PubKey, Provider: &c.Provider}

	return &c, nil
}
