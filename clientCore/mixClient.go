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

func (c *CryptoClient) GenerateDelaySequence(desiredRateParameter float64, length int) ([]float64, error){
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

func (c *CryptoClient) GetRandomMixSequence(mixes []config.MixPubs, length int) ([]config.MixPubs, error) {
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
