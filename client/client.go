package anonymous_messaging

import (
	packet "anonymous-messaging/packet_format"
	"fmt"
	"net"
	"os"
)

type Client struct {
	Id string
	Host string
	Port string
	PubKey int
	PrvKey int
}


type ClientOperations interface {
	EncodeMessage(message string) string
	DecodeMessage(message string) string
}

func (c Client) EncodeMessage(message string) packet.Packet {
	return packet.Encode(message, nil, nil)
}

func (c Client) DecodeMessage(packet string) string {
	return packet
}

func (c Client) SendMessage(message string, recipientHost string, recipientPort string) {
	packet := c.EncodeMessage(message)
	c.send(packet.ToString(), recipientHost, recipientPort)
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
