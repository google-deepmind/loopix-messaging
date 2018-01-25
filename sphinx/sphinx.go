package sphinx

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"strings"
	"anonymous-messaging/publics"
	"bytes"
	"errors"

	"github.com/protobuf/proto"
)

const (
	K = 16
	R = 5
	HEADERLENGTH = 192
	LAST_HOP_FLAG = "\xF0"
	RELAY_FLAG = "\xF1"

)

var curve = elliptic.P224()

func (p *SphinxPacket) Bytes() []byte{
	b, err := proto.Marshal(p)
	if err != nil {
		fmt.Println("Error in converting Packet to bytes ", err)
	}
	return b
}

func PacketFromBytes(bytes []byte) SphinxPacket {
	var packet SphinxPacket
	err := proto.Unmarshal(bytes, &packet)

	if err != nil {
		panic(err)
	}
	return packet
}

func (r *RoutingInfo) Bytes() []byte{
	b, err := proto.Marshal(r)
	if err != nil{
		fmt.Printf("Error during converting struct to bytes: %s", err)
	}
	return b
}


func RoutingInfoFromBytes(bytes []byte) RoutingInfo{

	var finalHopReconstruct RoutingInfo

	err := proto.Unmarshal(bytes, &finalHopReconstruct)
	if err != nil {
		panic(err)
	}
	return finalHopReconstruct
}


func PackForwardMessage(curve elliptic.Curve, path publics.E2EPath, delays []float64, message string) SphinxPacket{
	nodes := []publics.MixPubs{path.IngressProvider}
	nodes = append(nodes, path.Mixes...)
	nodes = append(nodes, path.EgressProvider)
	dest := path.Recipient

	asb, header := createHeader(curve, nodes, delays, dest)
	payload := encapsulateContent(asb, message)
	return SphinxPacket{Hdr: &header, Pld: payload}
}


func createHeader(curve elliptic.Curve, nodes []publics.MixPubs, delays []float64, dest publics.ClientPubs) ([]HeaderInitials, Header){

	x := randomBigInt(curve.Params())
	asb := getSharedSecrets(curve, nodes, x)
	computeFillers(nodes, asb)

	var commands []Commands
	for i, _ := range nodes {
		var c Commands
		if i == len(nodes) - 1 {
			c = Commands{Delay: delays[i], Flag: LAST_HOP_FLAG}
		} else {
			c = Commands{Delay: delays[i], Flag: RELAY_FLAG}
		}
		commands = append(commands, c)
	}

	header := encapsulateHeader(asb, nodes, commands, dest)
	return asb, header

}


func encapsulateContent(asb []HeaderInitials, message string) []byte{

	var enc []byte
	enc = []byte(message)
	for i := len(asb) - 1; i >= 0; i-- {
		sharedKey := KDF(asb[i].SecretHash)
		enc = AES_CTR(sharedKey, enc)
	}
	return enc
}


func encapsulateHeader(asb []HeaderInitials, nodes []publics.MixPubs, commands []Commands, destination publics.ClientPubs) Header{

	finalHop := RoutingInfo{NextHop: &Hop{Id: destination.Id, Address: destination.Host + ":" + destination.Port, PubKey: []byte{}}, RoutingCommands: &commands[len(commands) - 1], NextHopMetaData: []byte{}, Mac: []byte{}}

	encFinalHop := AES_CTR(KDF(asb[len(asb)-1].SecretHash), finalHop.Bytes())
	mac := computeMac(KDF(asb[len(asb)-1].SecretHash) , encFinalHop)

	routingCommands := [][]byte{encFinalHop}

	var encRouting []byte
	for i := len(nodes)-2; i >= 0; i-- {
		nextNode := nodes[i+1]
		routing := RoutingInfo{NextHop: &Hop{Id: nextNode.Id, Address: nextNode.Host+":"+nextNode.Port, PubKey: nodes[i+1].PubKey}, RoutingCommands: &commands[i], NextHopMetaData: routingCommands[len(routingCommands)-1], Mac: mac}

		encKey := KDF(asb[i].SecretHash)
		encRouting = AES_CTR(encKey, routing.Bytes())

		routingCommands = append(routingCommands, encRouting)
		mac = computeMac(KDF(asb[i].SecretHash) , encRouting)

	}
	return Header{Alpha: asb[0].Alpha, Beta: encRouting, Mac : mac}

}


func computeMac(key, data []byte) []byte{
	return Hmac(key, data)
}


func getSharedSecrets(curve elliptic.Curve, nodes []publics.MixPubs, initialVal big.Int) []HeaderInitials{

	blindFactors := []big.Int{initialVal}
	var tuples []HeaderInitials

	for _, n := range nodes {

		alpha := expo_group_base(curve, blindFactors)

		s := expo(n.PubKey, blindFactors)
		aes_s := KDF(s)

		blinder := computeBlindingFactor(curve, aes_s)
		blindFactors = append(blindFactors, *blinder)

		tuples = append(tuples, HeaderInitials{Alpha:alpha, Secret: s, Blinder: blinder.Bytes(), SecretHash: aes_s})
	}
	return tuples

}


func computeFillers(nodes []publics.MixPubs, tuples []HeaderInitials) string{

	filler := ""
	minLen := HEADERLENGTH - 32
	for i := 1; i < len(nodes); i++ {
		base := filler + strings.Repeat("\x00", K)
		kx := computeSharedSecretHash(tuples[i-1].SecretHash, []byte("hrhohrhohrhohrho"))
		mx := strings.Repeat("\x00", minLen) + base

		filler = BytesToString(AES_CTR([]byte(kx), []byte(mx)))
		filler = filler[minLen:]

		minLen = minLen - K
	}

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
	//order := curve.P
	nBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		panic(err)
	}
	return *nBig
}


func expo(base []byte, exp []big.Int) []byte{
	x := exp[0]
	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}

	baseX, baseY := elliptic.Unmarshal(elliptic.P224(), base)
	resultX, resultY := curve.Params().ScalarMult(baseX, baseY, x.Bytes())
	return elliptic.Marshal(curve, resultX, resultY)
}


func expo_group_base(curve elliptic.Curve, exp []big.Int) []byte{
	x := exp[0]

	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}

	resultX, resultY := curve.Params().ScalarBaseMult(x.Bytes())
	return elliptic.Marshal(curve, resultX, resultY)

}


func ProcessSphinxPacket(packetBytes []byte, privKey []byte) (Hop, Commands, SphinxPacket, error) {

	packet := PacketFromBytes(packetBytes)

	hop, commands, newHeader, err := ProcessSphinxHeader(*packet.Hdr, privKey)
	if err != nil {
		return Hop{}, Commands{}, SphinxPacket{}, err
	}
	newPayload, err := ProcessSphinxPayload(packet.Hdr.Alpha, packet.Pld, privKey)

	if err != nil {
		return Hop{}, Commands{}, SphinxPacket{}, err
	}
	return hop, commands, SphinxPacket{Hdr: &newHeader, Pld: newPayload}, nil
}


func ProcessSphinxHeader(packet Header, privKey []byte) (Hop, Commands, Header, error) {

	alpha := packet.Alpha
	beta := packet.Beta
	mac := packet.Mac

	curve := elliptic.P224()
	alphaX, alphaY := elliptic.Unmarshal(curve, alpha)
	sharedSecretX, sharedSecretY:= curve.Params().ScalarMult(alphaX, alphaY, privKey)
	sharedSecret := elliptic.Marshal(curve, sharedSecretX, sharedSecretY)

	aes_s := KDF(sharedSecret)
	encKey := KDF(aes_s)


	recomputedMac := computeMac(KDF(aes_s) , beta)

	if bytes.Compare(recomputedMac, mac) != 0 {
		return Hop{}, Commands{}, Header{}, errors.New("packet processing error: MACs are not matching")
	}

	blinder := computeBlindingFactor(curve, aes_s)
	newAlphaX, newAlphaY := curve.Params().ScalarMult(alphaX, alphaY, blinder.Bytes())
	newAlpha := elliptic.Marshal(curve, newAlphaX, newAlphaY)

	decBeta := AES_CTR(encKey, beta)
	nextHop, commands, nextBeta, nextMac := readBeta(RoutingInfoFromBytes(decBeta))

	return nextHop, commands, Header{Alpha: newAlpha, Beta: nextBeta, Mac: nextMac}, nil
}


func readBeta(beta RoutingInfo) (Hop, Commands, []byte, []byte){
	nextHop := *beta.NextHop
	commands := *beta.RoutingCommands
	nextBeta := beta.NextHopMetaData
	nextMac := beta.Mac

	return nextHop, commands, nextBeta, nextMac
}


func ProcessSphinxPayload(alpha []byte, payload []byte, privKey []byte) ([]byte, error) {

	curve := elliptic.P224()
	alphaX, alphaY := elliptic.Unmarshal(curve, alpha)
	sharedSecretX, sharedSecretY:= curve.Params().ScalarMult(alphaX, alphaY, privKey)
	sharedSecret := elliptic.Marshal(curve, sharedSecretX, sharedSecretY)

	aes_s := KDF(sharedSecret)
	decKey := KDF(aes_s)

	decPayload := AES_CTR(decKey, payload)

	return decPayload, nil
}