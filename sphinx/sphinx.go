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

/*	Package sphinx implements the library of a cryptographic packet format,
	which can be used to secure the content as well as the metadata of the transported
    messages.
*/

package sphinx

import (
	"anonymous-messaging/config"
	"anonymous-messaging/logging"

	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"

	"github.com/protobuf/proto"

	"bytes"
	"errors"
	"math/big"
	"strings"
)

var curve = elliptic.P224()
var logLocal = logging.PackageLogger()

const (
	K            = 16
	R            = 5
	headerLength = 192
	lastHopFlag  = "\xf0"
	relayFlag    = "\xf1"
)

// PackForwardMessage encapsulates the given message into the cryptographic Sphinx packet format.
// As arguments the function takes the path, consisting of the sequence of nodes the packet should traverse
// and the destination of the message, a set of delays and the information about the curve used to perform cryptographic
// operations.
// In order to encapsulate the message PackForwardMessage computes two parts of the packet - the header and
// the encrypted payload. If creating of any of the packet block failed, an error is returned. Otherwise,
// a Sphinx packet format is returned.
func PackForwardMessage(curve elliptic.Curve, path config.E2EPath, delays []float64, message string) (SphinxPacket, error) {
	nodes := []config.MixConfig{path.IngressProvider}
	nodes = append(nodes, path.Mixes...)
	nodes = append(nodes, path.EgressProvider)
	dest := path.Recipient

	asb, header, err := createHeader(curve, nodes, delays, dest)
	if err != nil {
		logLocal.WithError(err).Error("Error in PackForwardMessage - createHeader failed")
		return SphinxPacket{}, err
	}

	payload, err := encapsulateContent(asb, message)
	if err != nil {
		logLocal.WithError(err).Error("Error in PackForwardMessage - encapsulateContent failed")
		return SphinxPacket{}, err
	}
	return SphinxPacket{Hdr: &header, Pld: payload}, nil
}

// createHeader builds the Sphinx packet header, consisting of three parts: the public element, the encapsulated routing information
// and the message authentication code. createHeader layer encapsulates the routing information for each given node. The routing information
// contains information where the packet should be forwarded next, how long it should be delayed by the node, and if relevant additional
// auxiliary information. The message authentication code allows to detect tagging attacks.
// createHeader computes the secret shared key between sender and the nodes and destination, which are used as keys for encryption.
// createHeader returns the header and a list of the initial elements, used for creating the header. If any operation was unsuccessful
// createHeader returns an error.
func createHeader(curve elliptic.Curve, nodes []config.MixConfig, delays []float64, dest config.ClientConfig) ([]HeaderInitials, Header, error) {

	x, err := randomBigInt(curve.Params())

	if err != nil {
		logLocal.WithError(err).Error("Error in createHeader - randomBigInt failed")
		return nil, Header{}, err
	}

	asb, err := getSharedSecrets(curve, nodes, x)
	if err != nil {
		logLocal.WithError(err).Error("Error in createHeader - getSharedSecrets failed")
		return nil, Header{}, err
	}

	if len(asb) != len(nodes) {
		logLocal.WithError(err).Error("Error in createHeader - wrong number of shared secrets failed")
		return nil, Header{}, errors.New(" the number of shared secrets should be the same as the number of traversed nodes")
	}

	var commands []Commands
	for i, _ := range nodes {
		var c Commands
		if i == len(nodes)-1 {
			c = Commands{Delay: delays[i], Flag: lastHopFlag}
		} else {
			c = Commands{Delay: delays[i], Flag: relayFlag}
		}
		commands = append(commands, c)
	}

	header, err := encapsulateHeader(asb, nodes, commands, dest)
	if err != nil {
		logLocal.WithError(err).Error("Error in createHeader - encapsulateHeader failed")
		return nil, Header{}, err
	}
	return asb, header, nil

}

// encapsulateHeader layer encrypts the meta-data of the packet, containing information about the
// sequence of nodes the packet should traverse before reaching the destination, and message authentication codes,
// given the pre-computed shared keys which are used for encryption.
// encapsulateHeader returns the Header, or an error if any internal cryptographic of parsing operation failed.
func encapsulateHeader(asb []HeaderInitials, nodes []config.MixConfig, commands []Commands, destination config.ClientConfig) (Header, error) {
	finalHop := RoutingInfo{NextHop: &Hop{Id: destination.Id, Address: destination.Host + ":" + destination.Port, PubKey: []byte{}}, RoutingCommands: &commands[len(commands)-1], NextHopMetaData: []byte{}, Mac: []byte{}}

	finalHopBytes, err := proto.Marshal(&finalHop)
	if err != nil {
		return Header{}, err
	}

	encFinalHop, err := AES_CTR(KDF(asb[len(asb)-1].SecretHash), finalHopBytes)
	if err != nil {
		logLocal.WithError(err).Error("Error in encapsulateHeader - AES_CTR encryption failed")
		return Header{}, err
	}

	mac := computeMac(KDF(asb[len(asb)-1].SecretHash), encFinalHop)

	routingCommands := [][]byte{encFinalHop}

	var encRouting []byte
	for i := len(nodes) - 2; i >= 0; i-- {
		nextNode := nodes[i+1]
		routing := RoutingInfo{NextHop: &Hop{Id: nextNode.Id, Address: nextNode.Host + ":" + nextNode.Port, PubKey: nodes[i+1].PubKey}, RoutingCommands: &commands[i], NextHopMetaData: routingCommands[len(routingCommands)-1], Mac: mac}

		encKey := KDF(asb[i].SecretHash)
		routingBytes, err := proto.Marshal(&routing)

		if err != nil {
			return Header{}, err
		}

		encRouting, err = AES_CTR(encKey, routingBytes)
		if err != nil {
			return Header{}, err
		}

		routingCommands = append(routingCommands, encRouting)
		mac = computeMac(KDF(asb[i].SecretHash), encRouting)

	}
	return Header{Alpha: asb[0].Alpha, Beta: encRouting, Mac: mac}, nil

}

// encapsulateContent layer encrypts the given messages using a set of shared keys
// and the AES_CTR encryption.
// encapsulateContent returns the encrypted payload in byte representation. If the AES_CTR
// encryption failed encapsulateContent returns an error.
func encapsulateContent(asb []HeaderInitials, message string) ([]byte, error) {

	enc := []byte(message)
	err := error(nil)
	for i := len(asb) - 1; i >= 0; i-- {
		sharedKey := KDF(asb[i].SecretHash)
		enc, err = AES_CTR(sharedKey, enc)
		if err != nil {
			logLocal.WithError(err).Error("Error in encapsulateContent - AES_CTR encryption failed")
			return nil, err
		}

	}
	return enc, nil
}

// getSharedSecrets computes a sequence of HeaderInitial values, containing the initial elements,
// shared secrets and blinding factors for each node on the path. As input getSharedSecrets takes the initial
// secret value, the list of nodes, and the curve in which the cryptographic operations are performed.
// getSharedSecrets returns the list of computed HeaderInitials or an error.
func getSharedSecrets(curve elliptic.Curve, nodes []config.MixConfig, initialVal big.Int) ([]HeaderInitials, error) {

	blindFactors := []big.Int{initialVal}
	var tuples []HeaderInitials

	for _, n := range nodes {

		alpha := expoGroupBase(curve, blindFactors)

		s := expo(n.PubKey, blindFactors)
		aes_s := KDF(s)

		blinder, err := computeBlindingFactor(curve, aes_s)
		if err != nil {
			logLocal.WithError(err).Error("Error in getSharedSecrets - computeBlindingFactor failed")
			return nil, err
		}

		blindFactors = append(blindFactors, *blinder)
		tuples = append(tuples, HeaderInitials{Alpha: alpha, Secret: s, Blinder: blinder.Bytes(), SecretHash: aes_s})
	}
	return tuples, nil

}

// TODO: computeFillers needs to be fixed
func computeFillers(nodes []config.MixConfig, tuples []HeaderInitials) (string, error) {

	filler := ""
	minLen := headerLength - 32
	for i := 1; i < len(nodes); i++ {
		base := filler + strings.Repeat("\x00", K)
		kx, err := computeSharedSecretHash(tuples[i-1].SecretHash, []byte("hrhohrhohrhohrho"))
		if err != nil {
			return "", err
		}
		mx := strings.Repeat("\x00", minLen) + base

		xorVal, err := AES_CTR([]byte(kx), []byte(mx))
		if err != nil {
			logLocal.WithError(err).Error("Error in computeFillers - AES_CTR failed")
			return "", err
		}

		filler = BytesToString(xorVal)
		filler = filler[minLen:]

		minLen = minLen - K
	}

	return filler, nil

}

// computeBlindingFactor computes the blinding factor extracted from the
// shared secrets. Blinding factors allow both the sender and intermediate nodes
// recompute the shared keys used at each hop of the message processing.
// computeBlindingFactor returns a value of a blinding factor or an error.
func computeBlindingFactor(curve elliptic.Curve, key []byte) (*big.Int, error) {
	iv := []byte("initialvector000")
	blinderBytes, err := computeSharedSecretHash(key, iv)

	if err != nil {
		logLocal.WithError(err).Error("Error in computeBlindingFactor - computeSharedSecretHash failed")
		return &big.Int{}, err
	}

	return bytesToBigNum(curve, blinderBytes), nil
}

// computeSharedSecretHash computes the hash value of the shared secret key
// using AES_CTR.
func computeSharedSecretHash(key []byte, iv []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key)

	if err != nil {
		logLocal.WithError(err).Error("Error in computeSharedSecretHash - creating new AES cipher failed")
		return nil, err
	}

	stream := cipher.NewCTR(aesCipher, iv)
	plaintext := []byte("0000000000000000")

	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// ProcessSphinxPacket processes the sphinx packet using the given private key.
// ProcessSphinxPacket unwraps one layer of both the header and the payload encryption.
// ProcessSphinxPacket returns a new packet and the routing information which should
// be used by the processing node. If any cryptographic or parsing operation failed ProcessSphinxPacket
// returns an error.
func ProcessSphinxPacket(packetBytes []byte, privKey []byte) (Hop, Commands, []byte, error) {

	var packet SphinxPacket
	err := proto.Unmarshal(packetBytes, &packet)

	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxPacket - unmarshal of packet failed")
		return Hop{}, Commands{}, nil, err
	}

	hop, commands, newHeader, err := ProcessSphinxHeader(*packet.Hdr, privKey)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxPacket - ProcessSphinxHeader failed")
		return Hop{}, Commands{}, nil, err
	}

	newPayload, err := ProcessSphinxPayload(packet.Hdr.Alpha, packet.Pld, privKey)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxPacket - ProcessSphinxPayload failed")
		return Hop{}, Commands{}, nil, err
	}

	newPacket := SphinxPacket{Hdr: &newHeader, Pld: newPayload}
	newPacketBytes, err := proto.Marshal(&newPacket)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxPacket - marshal of packet failed")
		return Hop{}, Commands{}, nil, err
	}

	return hop, commands, newPacketBytes, nil
}

// ProcessSphinxHeader unwraps one layer of encryption from the header of a sphinx packet.
// ProcessSphinxHeader recomputes the shared key and checks whether the message authentication code is valid.
// If not, the packet is dropped and error is returned. If MAC checking was passed successfully ProcessSphinxHeader
// performs the AES_CTR decryption, recomputes the blinding factor and updates the init public element from the header.
// Next, ProcessSphinxHeader extracts the routing information from the decrypted packet and returns it, together with the
// updated init public element. If any crypto or parsing operation failed ProcessSphinxHeader returns an error.
func ProcessSphinxHeader(packet Header, privKey []byte) (Hop, Commands, Header, error) {

	alpha := packet.Alpha
	beta := packet.Beta
	mac := packet.Mac

	curve := elliptic.P224()
	alphaX, alphaY := elliptic.Unmarshal(curve, alpha)
	sharedSecretX, sharedSecretY := curve.Params().ScalarMult(alphaX, alphaY, privKey)
	sharedSecret := elliptic.Marshal(curve, sharedSecretX, sharedSecretY)

	aes_s := KDF(sharedSecret)
	encKey := KDF(aes_s)

	recomputedMac := computeMac(KDF(aes_s), beta)

	if bytes.Compare(recomputedMac, mac) != 0 {
		return Hop{}, Commands{}, Header{}, errors.New("packet processing error: MACs are not matching")
	}

	blinder, err := computeBlindingFactor(curve, aes_s)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxHeader - computeBlindingFactor failed")
		return Hop{}, Commands{}, Header{}, err
	}

	newAlphaX, newAlphaY := curve.Params().ScalarMult(alphaX, alphaY, blinder.Bytes())
	newAlpha := elliptic.Marshal(curve, newAlphaX, newAlphaY)

	decBeta, err := AES_CTR(encKey, beta)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxHeader - AES_CTR failed")
		return Hop{}, Commands{}, Header{}, err
	}

	var routingInfo RoutingInfo
	err = proto.Unmarshal(decBeta, &routingInfo)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxHeader - unmarshal of beta failed")
		return Hop{}, Commands{}, Header{}, err
	}
	nextHop, commands, nextBeta, nextMac := readBeta(routingInfo)

	return nextHop, commands, Header{Alpha: newAlpha, Beta: nextBeta, Mac: nextMac}, nil
}

// readBeta extracts all the fields from the RoutingInfo structure
func readBeta(beta RoutingInfo) (Hop, Commands, []byte, []byte) {
	nextHop := *beta.NextHop
	commands := *beta.RoutingCommands
	nextBeta := beta.NextHopMetaData
	nextMac := beta.Mac

	return nextHop, commands, nextBeta, nextMac
}

// ProcessSphinxPayload unwraps a single layer of the encryption from the sphinx packet payload.
// ProcessSphinxPayload first recomputes the shared secret which is used to perform the AES_CTR decryption.
// ProcessSphinxPayload returns the new packet payload or an error if the decryption failed.
func ProcessSphinxPayload(alpha []byte, payload []byte, privKey []byte) ([]byte, error) {

	curve := elliptic.P224()
	alphaX, alphaY := elliptic.Unmarshal(curve, alpha)
	sharedSecretX, sharedSecretY := curve.Params().ScalarMult(alphaX, alphaY, privKey)
	sharedSecret := elliptic.Marshal(curve, sharedSecretX, sharedSecretY)

	aes_s := KDF(sharedSecret)
	decKey := KDF(aes_s)

	decPayload, err := AES_CTR(decKey, payload)
	if err != nil {
		logLocal.WithError(err).Error("Error in ProcessSphinxPayload - AES_CTR decryption failed")
		return nil, err
	}

	return decPayload, nil
}
