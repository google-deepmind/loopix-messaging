/*
	Package clientCore implements all the necessary functions for the mix client, i.e., the core of the client
	which allows to process the received cryptographic packets.
 */

package clientCore

import (
	"anonymous-messaging/helpers"
	"anonymous-messaging/publics"
	sphinx "anonymous-messaging/new_packet_format"
	"crypto/elliptic"
)

type CryptoClient struct {
	Id     string
	PubKey publics.PublicKey
	PrvKey publics.PrivateKey
	Curve elliptic.Curve
}

func (c *CryptoClient) EncodeMessage(message string, path []publics.MixPubs, delays []float64, recipient publics.MixPubs) sphinx.SphinxPacket {
	var pubs []publics.PublicKey
	var commands []sphinx.Commands

	for _, v := range path {
		pubs = append(pubs, v.PubKey)
	}

	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}

	var packet sphinx.SphinxPacket
	packet = sphinx.PackForwardMessage(c.Curve, path, pubs, commands, recipient, message)

	return packet
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
