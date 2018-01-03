package clientCore

import (
	"anonymous-messaging/publics"
	"anonymous-messaging/packet_format"
	"anonymous-messaging/helpers"
	"time"
	"math/rand"
)

type MixClientIt interface {
	EncodeMessage(message string, path []publics.MixPubs, delays []float64) packet_format.Packet
	DecodeMessage(packet packet_format.Packet) packet_format.Packet
	GenerateDelaySequence(desiredRateParameter float64, length int) []float64
	GetRandomMixSequence(mixes []publics.MixPubs, length int) []publics.MixPubs
}

type MixClient struct {
	Id string
	PubKey int
	PrvKey int
}

func (c *MixClient) EncodeMessage(message string, path []publics.MixPubs, delays []float64) packet_format.Packet {
	return packet_format.Encode(message, path, delays)
}

func (c *MixClient) DecodeMessage(packet packet_format.Packet) packet_format.Packet {
	return packet_format.Decode(packet)
}

func (c *MixClient) GenerateDelaySequence(desiredRateParameter float64, length int) []float64 {
	rand.Seed(time.Now().UTC().UnixNano())

	var delays []float64
	for i := 0; i < length; i++{
		sample := rand.ExpFloat64() / desiredRateParameter
		delays = append(delays, sample)
	}
	return delays
}

func (c *MixClient) GetRandomMixSequence(mixes []publics.MixPubs, length int) []publics.MixPubs {
	if length > len(mixes) {
		return mixes
	} else {
		randomSeq := helpers.RandomSample(mixes, length)
		return randomSeq
	}
}

func NewMixClient(id string, pubKey, prvKey int) *MixClient{
	mixClient := MixClient{id, pubKey, prvKey}
	return &mixClient
}
