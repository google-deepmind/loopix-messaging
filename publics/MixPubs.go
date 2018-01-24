/*
	Package public implements struct for easy processing and storing of all public information
	of the network participants.
 */

package publics

import (
	"crypto/elliptic"
	"crypto/rand"
)

type MixPubs struct {
	Id     string
	Host   string
	Port   string
	PubKey []byte
}

func NewMixPubs(mixId, host, port string, pubKey []byte) MixPubs {
	mixPubs := MixPubs{Id: mixId, Host: host, Port: port, PubKey: pubKey}
	return mixPubs
}


type PublicKey struct {
	Value []byte
}

type PrivateKey struct {
	Value []byte
}

func GenerateKeyPair() ([]byte, []byte){
	priv, x, y, err  := elliptic.GenerateKey(elliptic.P224(), rand.Reader)

	if err != nil {
		panic(err)
	}

	return elliptic.Marshal(elliptic.P224(), x, y), priv
}