package tcpconn

import (
	"net"
	"fmt"
	"os"
)

type TCPClient struct {
	IP string
	Port string
}

func (client TCPClient) Connect(message string) {

	conn, err := net.Dial("tcp", client.IP + ":" + client.Port)
	defer conn.Close()

	if err != nil {
		fmt.Print("Error in Client connect", err.Error())
		os.Exit(1)
	}

	conn.Write([]byte(message))

	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	fmt.Println("Received answer: ", string(buff[:n]))


}