/*
	Package public implements struct for easy processing and storing of all public information
	of the network participants.
 */

package publics

import (
	"crypto/elliptic"
	"math/big"
)

type MixPubs struct {
	Id     string
	Host   string
	Port   string
	PubKey int64
}

func NewMixPubs(mixId, host, port string, pubKey int64) MixPubs {
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

type PrivateKey struct {
	Value []byte
}