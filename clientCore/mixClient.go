package anonymous_messaging

import (
	"anonymous-messaging/publics"
	"anonymous-messaging/packet_format"
	"time"
	"math/rand"
	"fmt"
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
	rand.Seed(time.Now().UTC().UnixNano())

	fmt.Println("Len: ", length)
	fmt.Println("Len of mixes: ", len(mixes))
	if length > len(mixes) {
		return mixes
	} else {
		permutedData := make([]publics.MixPubs, len(mixes))
		permutation := rand.Perm(len(mixes))

		for i, v := range permutation {
			permutedData[v] = mixes[i]
		}
		fmt.Println("Permuted: ", permutedData)

		fmt.Println("Cut: ", permutedData[:length])
		return permutedData[:length]
	}
}
