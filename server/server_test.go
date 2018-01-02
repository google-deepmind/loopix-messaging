package anonymous_messaging

import (
	"testing"
	"fmt"
	"os"

	"anonymous-messaging/packet_format"
)
var mixServer MixServer

func TestMain(m *testing.M) {
	mixServer = *NewMixServer("MixServer", "localhost", "9998", 1,0, "../pki/database.db")
	fmt.Println(mixServer)
	fmt.Println("HOST: ", mixServer.Host)
//	server.handleConnection(nil)

	code := m.Run()
	os.Exit(code)
}

func TestTest(t *testing.T){
	packet := packet_format.NewPacket("Hello", []float64{0.0, 0.0, 0.0}, nil, nil)
	mixServer.ReceivedPacket(packet)
}
