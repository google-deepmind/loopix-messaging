package clientCore

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"errors"

	"anonymous-messaging/config"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
)

var cryptoClient CryptoClient
var path config.E2EPath
var mixes []config.MixPubs

func Setup() error {
	for i := 0; i < 10; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil{
			return err
		}
		mixes = append(mixes, config.NewMixPubs(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub))
	}
	return nil
}


func TestMain(m *testing.M) {

	err := Setup()
	if err != nil {
		panic(m)
	}

	pubC, privC, err := sphinx.GenerateKeyPair()
	pub1, _, _ := sphinx.GenerateKeyPair()
	pub2, _, _ := sphinx.GenerateKeyPair()
	pubP, _, _ := sphinx.GenerateKeyPair()
	pubD, _, _ := sphinx.GenerateKeyPair()
	if err != nil{
		panic(err)
	}

	cryptoClient = CryptoClient{Id: "MixClient", PubKey: pubC, PrvKey: privC, Curve: elliptic.P224()}

	m1 := config.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	provider := config.MixPubs{Id: "Provider", Host: "localhost", Port: "3331", PubKey: pubP}
	recipient := config.ClientPubs{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD, Provider: &provider}

	path = config.E2EPath{IngressProvider: provider, Mixes: []config.MixPubs{m1, m2}, EgressProvider: provider, Recipient: recipient}

	os.Exit(m.Run())
}

func TestCryptoClient_EncodeMessage(t *testing.T) {

	message := "Hello world"
	delays := []float64{1.4, 2.5, 2.3, 3.5, 6.7}

	var commands []sphinx.Commands
	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}
	encoded, err := cryptoClient.EncodeMessage(message, path, delays)
	if err != nil{
		t.Error(err)
	}

	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(encoded))

}

func TestCryptoClient_DecodeMessage(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: &sphinx.Header{}, Pld: []byte("Message")}

	decoded, err := cryptoClient.DecodeMessage(packet)
	if err != nil{
		t.Fatal(err)
	}
	expected := packet
	assert.Equal(t, expected, decoded)
}

func TestCryptoClient_GenerateDelaySequence_Pass(t *testing.T) {
	delays, err := cryptoClient.GenerateDelaySequence(100, 5)
	if err != nil{
		t.Fatal(err)
	}
	assert.Equal(t, len(delays), 5, "The length of returned delays should be equal to theinput length")
	assert.Equal(t, reflect.TypeOf([]float64{}), reflect.TypeOf(delays), "The delays should be in float64 type")
}

func TestCryptoClient_GenerateDelaySequence_Fail(t *testing.T) {
	_, err := cryptoClient.GenerateDelaySequence(0, 5)
	assert.EqualError(t, errors.New("the parameter of exponential distribution has to be larger than zero"), err.Error(), "")
}

func Test_GetRandomMixSequence_TooFewMixes(t *testing.T) {

	sequence, err := cryptoClient.GetRandomMixSequence(mixes, 20)
	if err != nil{
		t.Fatal(err)
	}
	assert.Equal(t, 10, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")
}

func Test_GetRandomMixSequence_MoreMixes(t *testing.T) {

	sequence, err := cryptoClient.GetRandomMixSequence(mixes, 3)
	if err != nil{
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")

}

func Test_GetRandomMixSequence_FailEmptyList(t *testing.T) {
	_, err := cryptoClient.GetRandomMixSequence([]config.MixPubs{}, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}

func Test_GetRandomMixSequence_FailNonList(t *testing.T) {
	_, err := cryptoClient.GetRandomMixSequence(nil, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}