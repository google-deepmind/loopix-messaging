package clientCore

import (
	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"

	"github.com/stretchr/testify/assert"

	"crypto/elliptic"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

var cryptoClient CryptoClient
var path config.E2EPath
var mixes []config.MixConfig

func Setup() error {
	for i := 0; i < 10; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil {
			return err
		}
		mixes = append(mixes, config.NewMixConfig(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub))
	}

	// Create a mixClient
	pubC, privC, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	cryptoClient = CryptoClient{Id: "MixClient", PubKey: pubC, PrvKey: privC, Curve: elliptic.P224()}

	//Client a pair of mix configs, a single provider and a recipient
	pub1, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	pub2, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	m1 := config.MixConfig{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}

	pubP, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	pubD, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	provider := config.MixConfig{Id: "Provider", Host: "localhost", Port: "3331", PubKey: pubP}
	recipient := config.ClientConfig{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD, Provider: &provider}

	// Creating a test path
	path = config.E2EPath{IngressProvider: provider, Mixes: []config.MixConfig{m1, m2}, EgressProvider: provider, Recipient: recipient}

	return nil
}

func TestMain(m *testing.M) {

	err := Setup()
	if err != nil {
		panic(m)

	}
	os.Exit(m.Run())
}

func TestCryptoClient_EncodeMessage(t *testing.T) {

	delays := []float64{1.4, 2.5, 2.3, 3.5, 6.7}

	var commands []sphinx.Commands
	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}
	encoded, err := cryptoClient.EncodeMessage("Hello world", path, delays)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(encoded))

}

func TestCryptoClient_DecodeMessage(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: &sphinx.Header{}, Pld: []byte("Message")}

	decoded, err := cryptoClient.DecodeMessage(packet)
	if err != nil {
		t.Fatal(err)
	}
	expected := packet
	assert.Equal(t, expected, decoded)
}

func TestCryptoClient_GenerateDelaySequence_Pass(t *testing.T) {
	delays, err := cryptoClient.generateDelaySequence(100, 5)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(delays), 5, "The length of returned delays should be equal to theinput length")
	assert.Equal(t, reflect.TypeOf([]float64{}), reflect.TypeOf(delays), "The delays should be in float64 type")
}

func TestCryptoClient_GenerateDelaySequence_Fail(t *testing.T) {
	_, err := cryptoClient.generateDelaySequence(0, 5)
	assert.EqualError(t, errors.New("the parameter of exponential distribution has to be larger than zero"), err.Error(), "")
}

func Test_GetRandomMixSequence_TooFewMixes(t *testing.T) {

	sequence, err := cryptoClient.getRandomMixSequence(mixes, 20)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")
}

func Test_GetRandomMixSequence_MoreMixes(t *testing.T) {

	sequence, err := cryptoClient.getRandomMixSequence(mixes, 3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")

}

func Test_GetRandomMixSequence_FailEmptyList(t *testing.T) {
	_, err := cryptoClient.getRandomMixSequence([]config.MixConfig{}, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}

func Test_GetRandomMixSequence_FailNonList(t *testing.T) {
	_, err := cryptoClient.getRandomMixSequence(nil, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}
