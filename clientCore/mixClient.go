// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
	Package clientCore implements all the necessary functions for the mix client, i.e., the core of the client
	which allows to process the received cryptographic packets.
*/

package clientCore

import (
	"anonymous-messaging/config"
	"anonymous-messaging/helpers"
	"anonymous-messaging/logging"
	sphinx "anonymous-messaging/sphinx"

	"github.com/protobuf/proto"

	"crypto/elliptic"
	"errors"
)

var logLocal = logging.PackageLogger()

type NetworkPKI struct {
	Mixes   []config.MixConfig
	Clients []config.ClientConfig
}

type MixClient interface {
	EncodeIntoSphinxPacket(message string, recipient config.ClientConfig) ([]byte, error)
	DecodeSphinxPacket(packet sphinx.SphinxPacket) (sphinx.SphinxPacket, error)
	GetPublicKey() []byte
}

type CryptoClient struct {
	pubKey   []byte
	prvKey   []byte
	curve    elliptic.Curve
	Provider config.MixConfig
	Network  NetworkPKI
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
func (c *CryptoClient) createSphinxPacket(message string, recipient config.ClientConfig) ([]byte, error) {

	path, err := c.buildPath(recipient)
	if err != nil {
		logLocal.WithError(err).Error("Error in CreateSphinxPacket - generating random path failed")
		return nil, err
	}

	delays, err := c.generateDelaySequence(desiredRateParameter, path.Len())
	if err != nil {
		logLocal.WithError(err).Error("Error in CreateSphinxPacket - generating sequence of delays failed")
		return nil, err
	}

	sphinxPacket, err := sphinx.PackForwardMessage(c.curve, path, delays, message)
	if err != nil {
		logLocal.WithError(err).Error("Error in CreateSphinxPacket - the pack procedure failed")
		return nil, err
	}

	return proto.Marshal(&sphinxPacket)
}

// buildPath builds a path containing the sender's provider,
// a sequence (of length pre-defined in a config file) of randomly
// selected mixes and the recipient's provider
func (c *CryptoClient) buildPath(recipient config.ClientConfig) (config.E2EPath, error) {
	mixSeq, err := c.getRandomMixSequence(c.Network.Mixes, pathLength)
	if err != nil {
		logLocal.WithError(err).Error("Error in buildPath - generating random mix path failed")
		return config.E2EPath{}, err
	}
	path := config.E2EPath{IngressProvider: c.Provider, Mixes: mixSeq, EgressProvider: *recipient.Provider, Recipient: recipient}
	return path, nil
}

// getRandomMixSequence generates a random sequence of given length from all possible mixes.
// If the list of all active mixes is empty or the given length is larger than the set of active mixes,
// an error is returned.
func (c *CryptoClient) getRandomMixSequence(mixes []config.MixConfig, length int) ([]config.MixConfig, error) {
	if len(mixes) == 0 || mixes == nil {
		return nil, errors.New("cannot take a mix sequence from an empty list")
	}
	if length > len(mixes) {
		return mixes, nil
	} else {
		randomSeq, err := helpers.RandomSample(mixes, length)
		if err != nil {
			logLocal.WithError(err).Error("Error in getRandomMixSequence - sampling procedure failed")
			return nil, err
		}
		return randomSeq, nil
	}
}

// generateDelaySequence generates a given length sequence of float64 values. Values are generated
// following the exponential distribution. generateDelaySequence returnes a sequence or an error
// if any of the values could not be generate.
func (c *CryptoClient) generateDelaySequence(desiredRateParameter float64, length int) ([]float64, error) {
	var delays []float64
	for i := 0; i < length; i++ {
		d, err := helpers.RandomExponential(desiredRateParameter)
		if err != nil {
			logLocal.WithError(err).Error("Error in generateDelaySequence - generating random exponential sample failed")
			return nil, err
		}
		delays = append(delays, d)
	}
	return delays, nil
}

// EncodeMessage encodes given message into the Sphinx packet format. EncodeMessage takes as inputs
// the message and the recipient's public configuration.
// EncodeMessage returns the byte representation of the packet or an error if the packet could not be created.
func (c *CryptoClient) EncodeMessage(message string, recipient config.ClientConfig) ([]byte, error) {

	packet, err := c.createSphinxPacket(message, recipient)
	if err != nil {
		logLocal.WithError(err).Error("Error in EncodeMessage - the pack procedure failed")
		return nil, err
	}
	return packet, err
}

// DecodeMessage decodes the received sphinx packet.
// TODO: this function is finished yet.
func (c *CryptoClient) DecodeMessage(packet sphinx.SphinxPacket) (sphinx.SphinxPacket, error) {
	return packet, nil
}

func (c *CryptoClient) GetPublicKey() []byte {
	return c.pubKey
}

func NewCryptoClient(pubKey, privKey []byte, curve elliptic.Curve, provider config.MixConfig, network NetworkPKI) *CryptoClient {
	return &CryptoClient{pubKey: pubKey, prvKey: privKey, curve: curve, Provider: provider, Network: network}
}
