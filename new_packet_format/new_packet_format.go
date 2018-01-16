package new_packet_format

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"strings"
	"encoding/json"
	"anonymous-messaging/publics"
	"bytes"
)

const (
	K = 16
	R = 5
	HEADERLENGTH = 192
)

// KDF, Path Hop, Node metadata, blind, routingInfo, hdr


type HeaderInitials struct {
	Alpha publics.PublicKey
	Secret publics.PublicKey
	Blinder big.Int
	SecretHash []byte
}

type Header struct {
	Alpha publics.PublicKey
	Beta RoutingInfo
	Mac []byte
}


type Hop struct {
	Id string
	Address string
	PubKey []byte

}

type RoutingInfo struct {
	NextHop Hop
	RoutingCommands Commands
	NextHopMetaData []byte
	Mac []byte
}

func (r *RoutingInfo) Bytes() []byte{
	b, err := json.Marshal(r)
	if err != nil{
		fmt.Printf("Error during converting struct to bytes: %s", err)
	}
	return b

}

func RoutingInfoFromBytes(bytes []byte) RoutingInfo{

	var finalHopReconstruct RoutingInfo

	err := json.Unmarshal(bytes, &finalHopReconstruct)
	if err != nil {
		panic(err)
	}
	return finalHopReconstruct
}

type Commands struct {
	Delay float64
	Flag string
}


func createHeader(curve elliptic.Curve, nodes []publics.MixPubs, pubs []publics.PublicKey, commands []Commands, dest []string) Header{

	x := randomBigInt(curve.Params())
	asb := getSharedSecrets(curve, pubs, x)
	computeFillers(pubs, asb)
	header := encapsulateHeader(asb, nodes, pubs, commands, dest)
	return header

}

func encapsulateHeader(asb []HeaderInitials, nodes []publics.MixPubs, pubs []publics.PublicKey, commands []Commands, destination []string) Header{

	finalHop := RoutingInfo{NextHop: Hop{Id: destination[0], Address: destination[1], PubKey: []byte{}}, RoutingCommands: commands[len(commands) - 1], NextHopMetaData: nil, Mac: []byte{}}
	mac := computeMac(KDF(asb[len(asb)-1].SecretHash) , finalHop.Bytes())

	routingCommands := [][]byte{finalHop.Bytes()}

	var routing RoutingInfo
	for i := len(pubs)-2; i >= 0; i-- {
		nextNode := nodes[i+1]
		routing = RoutingInfo{NextHop: Hop{Id: nextNode.Id, Address: nextNode.Host+":"+nextNode.Port, PubKey: pubs[i+1].Bytes()}, RoutingCommands: commands[i], NextHopMetaData: routingCommands[len(routingCommands)-1], Mac: mac}

		//encKey := KDF(asb[i].SecretHash)
		//encRouting := AES_CTR(encKey, routing.Bytes())

		routingCommands = append(routingCommands, routing.Bytes())
		mac = computeMac(KDF(asb[i].SecretHash) , routing.Bytes())
		fmt.Println("Finished round : ", i)
	}
	return Header{Alpha: asb[0].Alpha, Beta: routing, Mac : mac}

}

func computeMac(key, data []byte) []byte{
	return Hmac(key, data)
}

func getSharedSecrets(curve elliptic.Curve, pubs []publics.PublicKey, initialVal big.Int) []HeaderInitials{

	blindFactors := []big.Int{initialVal}
	var tuples []HeaderInitials

	for _, key := range pubs {

		alpha := expo_group_base(curve, blindFactors)

		s := expo(key, blindFactors)
		aes_s := KDF(s.Bytes())

		blinder := computeBlindingFactor(curve, aes_s)
		blindFactors = append(blindFactors, *blinder)

		tuples = append(tuples, HeaderInitials{Alpha:alpha, Secret: s, Blinder: *blinder, SecretHash: aes_s})
	}
	return tuples

}

func computeFillers(pubs []publics.PublicKey, tuples []HeaderInitials) string{

	filler := ""
	minLen := HEADERLENGTH - 32
	for i := 1; i < len(pubs); i++ {
		base := filler + strings.Repeat("\x00", K)
		kx := computeSharedSecretHash(tuples[i-1].SecretHash, []byte("hrhohrhohrhohrho"))
		mx := strings.Repeat("\x00", minLen) + base

		filler = BytesToString(AES_CTR([]byte(kx), []byte(mx)))
		filler = filler[minLen:]

		minLen = minLen - K
	}

	fmt.Println("Filler len: ", len(filler))

	return filler

}

//func extractSecrets(tuples []HeaderInitials) []publics.PublicKey{
//
//	var secrets []publics.PublicKey
//	for _, v := range tuples {
//		secrets = append(secrets, v.Secret)
//	}
//	return secrets
//}


func computeBlindingFactor(curve elliptic.Curve, key []byte) *big.Int{
	iv := []byte("initialvector000")
	blinderBytes := computeSharedSecretHash(key, iv)

	return bytesToBigNum(curve, blinderBytes)
}

func computeSharedSecretHash(key []byte, iv []byte) []byte{
	aesCipher, err := aes.NewCipher(key)

	if err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(aesCipher, iv)
	plaintext := []byte("0000000000000000")

	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext
}

func KDF(key []byte) []byte{
	return hash(key)[:K]
}


func bytesToBigNum(curve elliptic.Curve, value []byte) *big.Int{
	nBig := new(big.Int)
	nBig.SetBytes(value)

	return new(big.Int).Mod(nBig, curve.Params().P)
}


func randomBigInt(curve *elliptic.CurveParams) big.Int{
	order := curve.P
	fmt.Println(order)

	nBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		panic(err)
	}
	return *nBig
}

func expo(base publics.PublicKey, exp []big.Int) publics.PublicKey{
	x := exp[0]
	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}
	curve := base.Curve
	resultX, resultY := curve.Params().ScalarMult(base.X, base.Y, x.Bytes())
	return publics.PublicKey{curve, resultX, resultY}
}

func expo_group_base(curve elliptic.Curve, exp []big.Int) publics.PublicKey{
	x := exp[0]

	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}

	resultX, resultY := curve.Params().ScalarBaseMult(x.Bytes())
	return publics.PublicKey{Curve: curve, X: resultX, Y: resultY}

}


func processSphinxPacket(packet Header, privKey publics.PrivateKey) Header{

	// make this function return an error when mac does not match

	alpha := packet.Alpha
	beta := packet.Beta
	mac := packet.Mac

	curve := elliptic.P224()
	sharedSecretX, sharedSecretY:= curve.Params().ScalarMult(alpha.X, alpha.Y, privKey.Value)
	sharedSecret := publics.PublicKey{elliptic.P224(), sharedSecretX, sharedSecretY}


	aes_s := KDF(sharedSecret.Bytes())
	encKey := KDF(aes_s)
	fmt.Println("KEY FOR DEC: ", encKey)

	// for new alpha
	blinder := computeBlindingFactor(curve, aes_s)
	new_alphaX, new_alphaY := curve.Params().ScalarMult(alpha.X, alpha.Y, blinder.Bytes())
	new_alpha := publics.PublicKey{curve, new_alphaX, new_alphaY}


	recomputedMac := computeMac(aes_s, beta.Bytes())

	if bytes.Compare(recomputedMac, mac) != 0 {
		fmt.Println("MAC's are not matching")
		// return an error here
	}

	// add decryption

	_, _, nextBeta, nextMac := readBeta(beta)


	newHeader := Header{Alpha: new_alpha, Beta: nextBeta, Mac: nextMac}
	return newHeader

}

func readBeta(beta RoutingInfo) (Hop, Commands, RoutingInfo, []byte){
	nextHop := beta.NextHop
	commands := beta.RoutingCommands
	nextBetaBytes := beta.NextHopMetaData
	nextMac := beta.Mac

	nextBeta := RoutingInfoFromBytes(nextBetaBytes)
	return nextHop, commands, nextBeta, nextMac
}