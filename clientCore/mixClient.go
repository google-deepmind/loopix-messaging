/*
	Package clientCore implements all the necessary functions for the mix client, i.e., the core of the client
	which allows to process the received cryptographic packets.
 */

package clientCore

import (
	"anonymous-messaging/helpers"
	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
	"errors"
)

type CryptoClient struct {
	Id     string
	PubKey []byte
	PrvKey []byte
	Curve elliptic.Curve
	ActiveMixes  []config.MixConfig
	Provider config.MixConfig
}

const (
	desiredRateParameter = 5
	pathLength           = 2
)

// CreateSphinxPacket responsible for sending a real message. Takes as input the message string
// and the public information about the destination.
// The function generates a random path and a set of random values from exponential distribution.
// Given those values it triggers the encode function, which packs the message into the
// sphinx cryptographic packet format. Next, the encoded packet is combined with a
// flag signaling that this is a usual network packet, and passed to be send.
// The function returns an error if any issues occurred.

func (c *CryptoClient) CreateSphinxPacket(message string, recipient config.ClientConfig) ([]byte, error) {


	path, err := c.buildPath(recipient)
	if err != nil{
		return nil, err
	}

	delays, err := c.generateDelaySequence(desiredRateParameter, path.Len())
	if err != nil{
		return nil, err
	}

	sphinxPacket, err := c.EncodeMessage(message, path, delays)
	if err != nil{
		return nil, err
	}

	return sphinxPacket, nil
}

// Function build a path containing the sender's provider,
// a sequence (of length pre-defined in a config file) of randomly
// selected mixes and the recipient's provider
func (c *CryptoClient) buildPath(recipient config.ClientConfig) (config.E2EPath, error) {
	mixSeq, err := c.getRandomMixSequence(c.ActiveMixes, pathLength)
	if err != nil{
		return config.E2EPath{}, err
	}

	path := config.E2EPath{IngressProvider: c.Provider, Mixes: mixSeq, EgressProvider: *recipient.Provider, Recipient: recipient}
	return path, nil
}

func (c *CryptoClient) getRandomMixSequence(mixes []config.MixConfig, length int) ([]config.MixConfig, error) {
	if len(mixes) == 0 || mixes == nil {
		return nil, errors.New("cannot take a mix sequence from an empty list")
	}
	if length > len(mixes) {
		return mixes, nil
	} else {
		randomSeq, err := helpers.RandomSample(mixes, length)
		if err != nil{
			return nil, err
		}
		return randomSeq, nil
	}
}

func (c *CryptoClient) generateDelaySequence(desiredRateParameter float64, length int) ([]float64, error){
	var delays []float64
	for i := 0; i < length; i++ {
		d, err := helpers.RandomExponential(desiredRateParameter)
		if err != nil{
			return nil, err
		}
		delays = append(delays, d)
	}
	return delays, nil
}

func (c *CryptoClient) EncodeMessage(message string, path config.E2EPath, delays []float64) ([]byte, error) {

	var packet sphinx.SphinxPacket
	packet, err := sphinx.PackForwardMessage(c.Curve, path, delays, message)
	if err != nil{
		return nil, err
	}

	return packet.Bytes()
}

func (c *CryptoClient) DecodeMessage(packet sphinx.SphinxPacket) (sphinx.SphinxPacket, error) {
	return packet, nil
}


