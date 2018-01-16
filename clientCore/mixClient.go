/*
	Package clientCore implements all the necessary functions for the mix client, i.e., the core of the client
	which allows to process the received cryptographic packets.
 */

package clientCore

import (
	"anonymous-messaging/helpers"
	"anonymous-messaging/packet_format"
	"anonymous-messaging/publics"
)

type CryptoClient struct {
	Id     string
	PubKey publics.PublicKey
	PrvKey publics.PrivateKey
}

func (c *CryptoClient) EncodeMessage(message string, path []publics.MixPubs, delays []float64, recipient publics.MixPubs) packet_format.Packet {
	return packet_format.Encode(message, path, delays)
}

func (c *CryptoClient) DecodeMessage(packet packet_format.Packet) packet_format.Packet {
	return packet_format.Decode(packet)
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
