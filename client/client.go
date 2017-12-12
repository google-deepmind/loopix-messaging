package anonymous_messaging

import (
	packet_format "anonymous-messaging/packet_format"
	"fmt"
	"net"
	"os"
	"math/rand"
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
	ActiveMixes []string //[]srv.MixServer
}


type ClientOperations interface {
	EncodeMessage(message string) string
	DecodeMessage(message string) string
}

func (c Client) EncodeMessage(message string, path []string, delays []float64) packet_format.Packet {
	return packet_format.Encode(message, path, delays)
}

func (c Client) DecodeMessage(packet string) string {
	return packet
}

func (c Client) SendMessage(message string, recipientHost string, recipientPort string) {
	path := c.ActiveMixes
	delays := c.GenerateDelaySequence(pathLength)
	packet := c.EncodeMessage(message, path, delays)
	c.send(packet_format.ToString(packet), recipientHost, recipientPort)
}

func (c Client) GenerateDelaySequence(length int) []float64{
	var delays []float64
	for i := 0; i < length; i++{
		sample := rand.ExpFloat64() / desiredRateParameter
		delays = append(delays, sample)
	}
	return delays
}

func (c Client) send(packet string, host string, port string) {
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

func NewClient(id, host, port string, pubKey, prvKey int) Client{
	return Client{Id:id, Host:host, Port:port, PubKey:pubKey, PrvKey:prvKey}
}
