package client
import (
	"testing"
	"github.com/stretchr/testify/assert"
	"anonymous-messaging/publics"
	"anonymous-messaging/packet_format"
	"os"
)
//

var client Client
var mixPubs []publics.MixPubs

func TestMain(m *testing.M){
	client = *NewClient("Client", "localhost", "3332", "../pki/database.db", 0, 0 )

	code := m.Run()
	os.Exit(code)
}

func TestClientProcessPacket(t *testing.T){
	packet := packet_format.NewPacket("Message", []float64{0.1, 0.2, 0.3}, mixPubs, nil)

	m := client.ProcessPacket(packet)
	assert.Equal(t, m, "Message", "The final message should be the same as the init one")
}
