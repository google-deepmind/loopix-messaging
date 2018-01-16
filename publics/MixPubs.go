/*
	Package public implements struct for easy processing and storing of all public information
	of the network participants.
 */

package publics

import (
	"crypto/elliptic"
	"math/big"
	"crypto/rand"
)

type MixPubs struct {
	Id     string
	Host   string
	Port   string
	PubKey PublicKey
}

func NewMixPubs(mixId, host, port string, pubKey PublicKey) MixPubs {
	mixPubs := MixPubs{Id: mixId, Host: host, Port: port, PubKey: pubKey}
	return mixPubs
}

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

func (p *PublicKey) Bytes() []byte{
	return elliptic.Marshal(p.Curve, p.X, p.Y)
}

func PubKeyFromBytes(curve elliptic.Curve, keyBytes []byte) PublicKey{
	x, y := elliptic.Unmarshal(curve, keyBytes)
	return PublicKey{Curve: curve, X: x, Y: y}
}

type PrivateKey struct {
	Value []byte
}

func GenerateKeyPair() (PublicKey, PrivateKey){
	priv, x, y, err  := elliptic.GenerateKey(elliptic.P224(), rand.Reader)

	if err != nil {
		panic(err)
	}

	pubKey := PublicKey{elliptic.P224(), x, y}
	privKey := PrivateKey{priv}
	return pubKey, privKey
}