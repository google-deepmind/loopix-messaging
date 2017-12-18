package anonymous_messaging

import (
	packet_format "anonymous-messaging/packet_format"
	"fmt"
	"net"
	"os"
	"math/rand"
	"time"
	"anonymous-messaging/publics"
	"anonymous-messaging/pki"
	"github.com/jmoiron/sqlx"
)

const (
	desiredRateParameter = 5
	pathLength = 2
)

type Client struct {
	Id string
	Host string
	Port string
	PubKey int
	PrvKey int
	ActiveMixes []publics.MixPubs

	listener *net.TCPListener
}


type ClientOperations interface {
	EncodeMessage(message string) string
	DecodeMessage(message string) string
}

func (c Client) EncodeMessage(message string, path []publics.MixPubs, delays []float64) packet_format.Packet {
	return packet_format.Encode(message, path, delays)
}

func (c Client) DecodeMessage(packet packet_format.Packet) packet_format.Packet {
	return packet_format.Decode(packet)
}

func (c Client) SendMessage(message string, recipientHost string, recipientPort string) {
	path := c.GetRandomMixSequence(c.ActiveMixes, pathLength)
	path = append(path, publics.MixPubs{Id:c.Id, Host:c.Host, Port:c.Port, PubKey:0})
	delays := c.GenerateDelaySequence(desiredRateParameter, pathLength)
	packet := c.EncodeMessage(message, path, delays)
	c.Send(packet_format.ToString(packet), recipientHost, recipientPort)
}

func (c Client) GenerateDelaySequence(desiredRateParameter float64, length int) []float64{
	rand.Seed(time.Now().UTC().UnixNano())
	var delays []float64
	for i := 0; i < length; i++{
		sample := rand.ExpFloat64() / desiredRateParameter
		delays = append(delays, sample)
	}
	return delays
}

func (c Client) GetRandomMixSequence(data []publics.MixPubs, length int) []publics.MixPubs {
	rand.Seed(time.Now().UTC().UnixNano())
	permutedData := make([]publics.MixPubs, len(data))
	permutation := rand.Perm(len(data))
	for i, v := range permutation {
		permutedData[v] = data[i]
	}
	return permutedData[:length]
}

func (c Client) Send(packet string, host string, port string) {
	conn, err := net.Dial("tcp", host + ":" + port)
	defer conn.Close()

	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write([]byte(packet))

	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	fmt.Println("Received answer: ", string(buff[:n]))
}



func (c Client) listenForConnections() {
	for {
		conn, err := c.listener.Accept()

		if err != nil {
			fmt.Println("Error in connection accepting:", err.Error())
			os.Exit(1)
		}
		fmt.Println(conn)
		//fmt.Println("Received connection from : ", conn.RemoteAddr())
		go c.handleConnection(conn)
	}
}

func (c Client) handleConnection(conn net.Conn) {
	fmt.Println("> Handle Connection")

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)

	if err != nil {
		fmt.Println()
	}

	c.ProcessPacket(packet_format.FromString(string(buff[:reqLen])))
	conn.Close()
}

func (c Client) ProcessPacket(packet packet_format.Packet) string{
	fmt.Println("Processing packet")
	return packet.Message
}

func (c Client) Start() {

	defer c.Run()

	c.SaveInPKI()
	c.ReadInMixnetPKI()
}

func (c Client) Run() {
	fmt.Println("> Client is running")
	defer c.listener.Close()
	finish := make(chan bool)

	go func() {
		fmt.Println("Listening on " + c.Host + ":" + c.Port)
		c.listenForConnections()
	}()

	//c.SendMessage("Hello world", "localhost", "3330")
	<-finish
}

func (c Client) ReadInMixnetPKI() {
	fmt.Println("Reading network")

	db := c.ConnectToPKI()
	records := pki.QueryDatabase(db,"Mixes")

	for records.Next() {
		results := make(map[string]interface{})
		err := records.MapScan(results)

		if err != nil {
			panic(err)

		}

		pubs := publics.NewMixPubs(string(results["MixId"].([]byte)), string(results["Host"].([]byte)),
									string(results["Port"].([]byte)), results["PubKey"].(int64))

		c.ActiveMixes = append(c.ActiveMixes, pubs)
	}
	fmt.Println("> The mix network data is uploaded.")
}


func (c Client) SaveInPKI() {
	fmt.Println("> Saving Client Public Info into Database")

	db := c.ConnectToPKI()

	pubInfo := make(map[string]interface{})
	pubInfo["MixId"] = c.Id
	pubInfo["Host"] = c.Host
	pubInfo["Port"] = c.Port
	pubInfo["PubKey"] = c.PubKey
	pki.InsertToTable(db, "Clients", pubInfo)

	fmt.Println("> Public info of the mixserver saved in database")
}

func (c Client) ConnectToPKI() *sqlx.DB{
	db := pki.CreateAndOpenDatabase("./pki/database.db", "./pki/database.db", "sqlite3")

	params := make(map[string]string)
	params["ClientId"] = "TEXT"
	params["Host"] = "TEXT"
	params["Port"] = "TEXT"
	params["PubKey"] = "NUM"
	pki.CreateTable(db, "Clients", params)

	return db
}

func NewClient(id, host, port string, pubKey, prvKey int) Client{
	c := Client{Id:id, Host:host, Port:port, PubKey:pubKey, PrvKey:prvKey}

	addr, err := net.ResolveTCPAddr("tcp", c.Host + ":" + c.Port)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	c.listener, err = net.ListenTCP("tcp", addr)
	return c
}
