/*
	Package clientCore implements all the necessary functions for the mix client, i.e., the core of the client
	which allows to process the received cryptographic packets.
 */

package clientCore

import (
	"anonymous-messaging/helpers"
	"anonymous-messaging/publics"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
)

type CryptoClient struct {
	Id     string
	PubKey []byte //publics.PublicKey
	PrvKey []byte //publics.PrivateKey
	Curve elliptic.Curve
}

func (c *CryptoClient) EncodeMessage(message string, path publics.E2EPath, delays []float64) []byte {

	var packet sphinx.SphinxPacket
	packet = sphinx.PackForwardMessage(c.Curve, path, delays, message)

	return packet.Bytes()
}

func (c *CryptoClient) DecodeMessage(packet sphinx.SphinxPacket) sphinx.SphinxPacket {
	return packet
}

func (c *CryptoClient) GenerateDelaySequence(desiredRateParameter float64, length int) []float64 {
	var delays []float64
	for i := 0; i < length; i++ {
		delays = append(delays, helpers.RandomExponential(desiredRateParameter))
	}
	return delays
}

func (c *CryptoClient) GetRandomMixSequence(mixes []publics.MixPubs, length int) []publics.MixPubs {
	if length > len(mixes) {
		return mixes
	} else {
		randomSeq := helpers.RandomSample(mixes, length)
		return randomSeq
	}
}
