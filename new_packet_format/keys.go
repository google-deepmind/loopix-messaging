package new_packet_format

import (
	"crypto/elliptic"
	"math/big"
)

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}


func (p *PublicKey) Bytes() []byte{
	return elliptic.Marshal(p.Curve, p.X, p.Y)
}

type PrivateKey struct {
	privk []byte
}
