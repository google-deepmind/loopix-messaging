package anonymous_messaging
//import (
//
//	"testing"
//	"github.com/stretchr/testify/assert"
//	"reflect"
//)
//
//func TestClientEncode(t *testing.T) {
//
//	p := c.EncodeMessage("Hello world", mixPubs, []float64{1.4, 2.5, 2.3})
//	assert.Equal(t, p.Message, "Hello world", "The messages should be the same")
//	assert.Equal(t, p.Path, mixPubs, "The path should be the same")
//}
//
//func TestClientDecode(t *testing.T) {
//	d := c.DecodeMessage(packet)
//	if reflect.DeepEqual(d, packet) == false {
//		t.Error("Error in decode: ")
//	}
//}
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
//func TestClientProcessPacket(t *testing.T){
//	m := c.ProcessPacket(packet)
//	assert.Equal(t, m, "Hello you", "The final message should be the same as the init one")
//}
