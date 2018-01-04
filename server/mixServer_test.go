package server

import (
	"anonymous-messaging/packet_format"
	"fmt"
	"os"
	"testing"
)

var mixServer MixServer

func TestMain(m *testing.M) {
	mixServer = *NewMixServer("MixServer", "localhost", "9998", 1, 0, "../pki/database.db")
	fmt.Println(mixServer)

	code := m.Run()
	os.Exit(code)
}

func TestTest(t *testing.T) {
	packet := packet_format.NewPacket("Hello", []float64{0.0, 0.0, 0.0}, nil, nil)
	mixServer.ReceivedPacket(packet)

	//serverSide, clientSide := net.Pipe()
	//
	//go mixServer.HandleConnection(serverSide)
	//
	//msg := "message to server"
	//fmt.Println(msg)
	//clientSide.Write([]byte(msg))
	////
	//if err != nil {
	//	fmt.Printf("Unable to send message : %v\n", err)
	//}

}
