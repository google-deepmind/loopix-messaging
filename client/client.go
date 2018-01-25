/*
	Package client implements the class of a network client which can interact with a mix network.
*/

package client

import (
	"fmt"
	"net"
	"os"

	"anonymous-messaging/clientCore"
	"anonymous-messaging/networker"
	"anonymous-messaging/pki"
	"anonymous-messaging/publics"
	"github.com/jmoiron/sqlx"
	"crypto/elliptic"
)

const (
	desiredRateParameter = 5
	pathLength           = 2
)

type ClientIt interface {
	networker.NetworkClient
	networker.NetworkServer
	SendMessage(message string, recipient publics.MixPubs)
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

	ActiveMixes  []publics.MixPubs
	OtherClients []publics.ClientPubs

	listener *net.TCPListener

	pkiDir string
	Provider publics.MixPubs
}

func (c *Client) SendMessage(message string, recipient publics.ClientPubs) {
	mixSeq := c.GetRandomMixSequence(c.ActiveMixes, pathLength)

	var path publics.E2EPath
	path.IngressProvider = c.Provider
	path.Mixes = mixSeq
	path.EgressProvider = *recipient.Provider
	path.Recipient = recipient

	delays := c.GenerateDelaySequence(desiredRateParameter, path.Len())

	packet := c.EncodeMessage(message, path, delays)
	c.Send(packet, path.IngressProvider.Host, path.IngressProvider.Port)
}

func (c *Client) Send(packet []byte, host string, port string) {
	conn, err := net.Dial("tcp", host+":"+port)
	defer conn.Close()

	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write(packet)
}

func (c *Client) ListenForIncomingConnections() {
	for {
		conn, err := c.listener.Accept()

		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		go c.HandleConnection(conn)
	}
}

func (c *Client) HandleConnection(conn net.Conn) {
	fmt.Println("> Handle Connection")

	buff := make([]byte, 1024)

	reqLen, err := conn.Read(buff)
	fmt.Println(reqLen)
	if err != nil {
		fmt.Println()
	}

	c.ProcessPacket(buff[:reqLen])
	conn.Close()
}

func (c *Client) ProcessPacket(packet []byte) []byte {
	fmt.Println("Processing packet: ", packet)
	return packet
}

func (c *Client) Start() {

	defer c.Run()

	c.ReadInClientsPKI(c.pkiDir)
	c.ReadInMixnetPKI(c.pkiDir)

}

func (c *Client) contactProvider() {
	fmt.Println("Sending to provider")
}

func (c *Client) Run() {
	fmt.Println("> Client is running")

	defer c.listener.Close()
	finish := make(chan bool)

	go func() {
		fmt.Println("Listening on " + c.Host + ":" + c.Port)
		c.ListenForIncomingConnections()
	}()

	go func() {
		c.SendMessage("Hello world, this is me", c.OtherClients[0])
	}()

	<-finish
}

func (c *Client) ReadInMixnetPKI(pkiName string) {
	fmt.Println("Reading network from pki:  ", pkiName)

	db := c.ConnectToPKI(pkiName)
	records := pki.QueryDatabase(db, "Mixes")

	for records.Next() {
		results := make(map[string]interface{})
		err := records.MapScan(results)

		if err != nil {
			panic(err)

		}
		pubs := publics.NewMixPubs(string(results["MixId"].([]byte)), string(results["Host"].([]byte)),
			string(results["Port"].([]byte)), results["PubKey"].([]byte))

		c.ActiveMixes = append(c.ActiveMixes, pubs)
	}
	fmt.Println("> The mix network data is uploaded.")
}

func (c *Client) ReadInClientsPKI(pkiName string) {
	fmt.Println("Reading public information about clients")


	db := c.ConnectToPKI(pkiName)
	records := pki.QueryDatabase(db, "Clients")

	for records.Next() {
		results := make(map[string]interface{})
		err := records.MapScan(results)

		if err != nil {
			panic(err)

		}
		provider, err := publics.MixPubsFromBytes(results["Provider"].([]byte))
		if err != nil {
			panic(err)
		}
		pubs := publics.NewClientPubs(string(results["ClientId"].([]byte)), string(results["Host"].([]byte)),
			string(results["Port"].([]byte)), results["PubKey"].([]byte), provider)
		c.OtherClients = append(c.OtherClients, pubs)
	}
	fmt.Println("> The clients data is uploaded.")
}

func (c *Client) ConnectToPKI(dbName string) *sqlx.DB {
	return pki.OpenDatabase(dbName, "sqlite3")
}

func SaveInPKI(c Client, pkiDir string) {
	fmt.Println("> Saving Client Public Info into Database")

	db := pki.OpenDatabase(pkiDir, "sqlite3")

	params := make(map[string]string)
	params["ClientId"] = "TEXT"
	params["Host"] = "TEXT"
	params["Port"] = "TEXT"
	params["PubKey"] = "BLOB"
	params["Provider"] = "BLOB"
	pki.CreateTable(db, "Clients", params)

	pubInfo := make(map[string]interface{})
	pubInfo["ClientId"] = c.Id
	pubInfo["Host"] = c.Host
	pubInfo["Port"] = c.Port
	pubInfo["PubKey"] = c.PubKey
	pubInfo["Provider"], _ = publics.MixPubsToBytes(c.Provider)
	pki.InsertToTable(db, "Clients", pubInfo)

	fmt.Println("> Public info of the client saved in database")
	db.Close()
}

func NewClient(id, host, port string, pubKey []byte, prvKey []byte, pkiDir string, provider publics.MixPubs) *Client {
	core := clientCore.CryptoClient{Id: id, PubKey: pubKey, PrvKey: prvKey, Curve: elliptic.P224()}

	c := Client{Host: host, Port: port, CryptoClient: core, Provider: provider, pkiDir: pkiDir}
	SaveInPKI(c, pkiDir)

	addr, err := net.ResolveTCPAddr("tcp", c.Host+":"+c.Port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	c.listener, err = net.ListenTCP("tcp", addr)
	return &c
}
