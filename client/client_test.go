package anonymous_messaging
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
	m1 := publics.MixPubs{"Mix1", "localhost", "3330", 0}
	m2 := publics.MixPubs{"Mix2", "localhost", "3331", 0}

	mixPubs = []publics.MixPubs{m1, m2}

	code := m.Run()
	os.Exit(code)
}


func TestClientEncode(t *testing.T) {

	message := "Hello world"
	path := mixPubs
	delays := []float64{1.4, 2.5, 2.3}

	encoded := client.EncodeMessage(message, path, delays)
	expected := packet_format.Encode(message, path, delays)
	assert.Equal(t, encoded, expected, "The packets should be the same")
}

func TestClientDecode(t *testing.T) {

	packet := packet_format.NewPacket("Message", []float64{0.1, 0.2, 0.3}, mixPubs, nil)

	decoded := client.DecodeMessage(packet)
	expected := packet_format.Decode(packet)

	assert.Equal(t, decoded, expected, "The packets should be the same")
}
//
//func TestClientSendMessage(t *testing.T) {
//}
//
//func TestClientGenerateDelays(t *testing.T){
//	delays := c.GenerateDelaySequence(5,3)
//	if len(delays) != 3 {
//		t.Error("Incorrect number of generated delays")
//	}
//	if reflect.TypeOf(delays).Elem().Kind() != reflect.Float64 {
//		t.Error("Incorrect type of generated delays")
//	}
//}
//
func TestClientProcessPacket(t *testing.T){
	packet := packet_format.NewPacket("Message", []float64{0.1, 0.2, 0.3}, mixPubs, nil)

	m := client.ProcessPacket(packet)
	assert.Equal(t, m, "Message", "The final message should be the same as the init one")
}
