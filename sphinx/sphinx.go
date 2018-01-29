package sphinx

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
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

func (p *SphinxPacket) Bytes() ([]byte, error) {
	b, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func PacketFromBytes(bytes []byte) (SphinxPacket, error) {
	var packet SphinxPacket
	err := proto.Unmarshal(bytes, &packet)

	if err != nil {
		return SphinxPacket{}, err
	}

	return packet, nil
}

func (r *RoutingInfo) Bytes() ([]byte, error) {
	b, err := proto.Marshal(r)
	if err != nil{
		return nil, err
	}
	return b, nil
}


func RoutingInfoFromBytes(bytes []byte) (RoutingInfo, error) {

	var finalHopReconstruct RoutingInfo

	err := proto.Unmarshal(bytes, &finalHopReconstruct)

	if err != nil {
		return RoutingInfo{}, err
	}

	return finalHopReconstruct, nil
}


func PackForwardMessage(curve elliptic.Curve, path publics.E2EPath, delays []float64, message string) (SphinxPacket, error){
	nodes := []publics.MixPubs{path.IngressProvider}
	nodes = append(nodes, path.Mixes...)
	nodes = append(nodes, path.EgressProvider)
	dest := path.Recipient

	asb, header, err := createHeader(curve, nodes, delays, dest)
	if err != nil{
		return SphinxPacket{}, err
	}

	payload, err := encapsulateContent(asb, message)
	if err != nil{
		return SphinxPacket{}, err
	}

	return SphinxPacket{Hdr: &header, Pld: payload}, nil
}


func createHeader(curve elliptic.Curve, nodes []publics.MixPubs, delays []float64, dest publics.ClientPubs) ([]HeaderInitials, Header, error){

	x, err := randomBigInt(curve.Params())

	if err != nil {
		return nil, Header{}, err
	}

	asb, err := getSharedSecrets(curve, nodes, x)
	if err != nil{
		return nil, Header{}, err
	}


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

	header, err := encapsulateHeader(asb, nodes, commands, dest)
	if err !=nil{
		return nil, Header{}, err
	}
	return asb, header, nil

}


func encapsulateContent(asb []HeaderInitials, message string) ([]byte, error) {

	enc := []byte(message)
	err := error(nil)
	for i := len(asb) - 1; i >= 0; i-- {
		sharedKey := KDF(asb[i].SecretHash)
		enc, err = AES_CTR(sharedKey, enc)
		if err != nil{
			return nil, err
		}

	}
	return enc, nil
}


func encapsulateHeader(asb []HeaderInitials, nodes []publics.MixPubs, commands []Commands, destination publics.ClientPubs) (Header, error){

	finalHop := RoutingInfo{NextHop: &Hop{Id: destination.Id, Address: destination.Host + ":" + destination.Port, PubKey: []byte{}}, RoutingCommands: &commands[len(commands) - 1], NextHopMetaData: []byte{}, Mac: []byte{}}

	finalHopBytes, err := finalHop.Bytes()
	if err != nil{
		return Header{}, err
	}

	encFinalHop, err := AES_CTR(KDF(asb[len(asb)-1].SecretHash), finalHopBytes)
	if err != nil{
		return Header{}, err
	}

	mac := computeMac(KDF(asb[len(asb)-1].SecretHash) , encFinalHop)

	routingCommands := [][]byte{encFinalHop}

	var encRouting []byte
	for i := len(nodes)-2; i >= 0; i-- {
		nextNode := nodes[i+1]
		routing := RoutingInfo{NextHop: &Hop{Id: nextNode.Id, Address: nextNode.Host+":"+nextNode.Port, PubKey: nodes[i+1].PubKey}, RoutingCommands: &commands[i], NextHopMetaData: routingCommands[len(routingCommands)-1], Mac: mac}

		encKey := KDF(asb[i].SecretHash)
		routingBytes, err := routing.Bytes()

		if err != nil{
			return Header{}, err
		}

		encRouting, err = AES_CTR(encKey, routingBytes)
		if err != nil{
			return Header{}, err
		}

		routingCommands = append(routingCommands, encRouting)
		mac = computeMac(KDF(asb[i].SecretHash) , encRouting)

	}
	return Header{Alpha: asb[0].Alpha, Beta: encRouting, Mac : mac}, nil

}


func computeMac(key, data []byte) []byte{
	return Hmac(key, data)
}


func getSharedSecrets(curve elliptic.Curve, nodes []publics.MixPubs, initialVal big.Int) ([]HeaderInitials, error){

	blindFactors := []big.Int{initialVal}
	var tuples []HeaderInitials

	for _, n := range nodes {

		alpha := expo_group_base(curve, blindFactors)

		s := expo(n.PubKey, blindFactors)
		aes_s := KDF(s)

		blinder, err := computeBlindingFactor(curve, aes_s)
		if err != nil{
			return nil, err
		}

		blindFactors = append(blindFactors, *blinder)
		tuples = append(tuples, HeaderInitials{Alpha:alpha, Secret: s, Blinder: blinder.Bytes(), SecretHash: aes_s})
	}
	return tuples, nil

}


func computeFillers(nodes []publics.MixPubs, tuples []HeaderInitials) (string, error) {

	filler := ""
	minLen := HEADERLENGTH - 32
	for i := 1; i < len(nodes); i++ {
		base := filler + strings.Repeat("\x00", K)
		kx, err := computeSharedSecretHash(tuples[i-1].SecretHash, []byte("hrhohrhohrhohrho"))
		if err != nil{
			return "", err
		}
		mx := strings.Repeat("\x00", minLen) + base

		xorVal, err := AES_CTR([]byte(kx), []byte(mx))
		if err != nil{
			return "", err
		}

		filler = BytesToString(xorVal)
		filler = filler[minLen:]

		minLen = minLen - K
	}

	return filler, nil

}


func computeBlindingFactor(curve elliptic.Curve, key []byte) (*big.Int, error) {
	iv := []byte("initialvector000")
	blinderBytes, err := computeSharedSecretHash(key, iv)

	if err != nil{
		return &big.Int{}, err
	}

	return bytesToBigNum(curve, blinderBytes), nil
}


func computeSharedSecretHash(key []byte, iv []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(aesCipher, iv)
	plaintext := []byte("0000000000000000")

	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}


func KDF(key []byte) []byte{
	return hash(key)[:K]
}


func bytesToBigNum(curve elliptic.Curve, value []byte) *big.Int{
	nBig := new(big.Int)
	nBig.SetBytes(value)

	return new(big.Int).Mod(nBig, curve.Params().P)
}


func randomBigInt(curve *elliptic.CurveParams) (big.Int, error) {
	//order := curve.P
	nBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return big.Int{}, err
	}
	return *nBig, nil
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


func ProcessSphinxPacket(packetBytes []byte, privKey []byte) (Hop, Commands, []byte, error) {

	packet, err := PacketFromBytes(packetBytes)

	if err != nil {
		return Hop{}, Commands{}, nil, err
	}

	hop, commands, newHeader, err := ProcessSphinxHeader(*packet.Hdr, privKey)
	if err != nil {
		return Hop{}, Commands{}, nil, err
	}

	newPayload, err := ProcessSphinxPayload(packet.Hdr.Alpha, packet.Pld, privKey)
	if err != nil {
		return Hop{}, Commands{}, nil, err
	}

	newPacket := SphinxPacket{Hdr: &newHeader, Pld: newPayload}
	newPacketBytes, err := newPacket.Bytes()
	if err != nil{
		return Hop{}, Commands{}, nil, err
	}

	return hop, commands, newPacketBytes, nil
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

	blinder, err := computeBlindingFactor(curve, aes_s)
	if err != nil{
		return Hop{}, Commands{}, Header{}, err
	}

	newAlphaX, newAlphaY := curve.Params().ScalarMult(alphaX, alphaY, blinder.Bytes())
	newAlpha := elliptic.Marshal(curve, newAlphaX, newAlphaY)

	decBeta, err := AES_CTR(encKey, beta)
	if err != nil{
		return Hop{}, Commands{}, Header{}, err
	}

	routingInfo, err := RoutingInfoFromBytes(decBeta)
	if err != nil{
		return Hop{}, Commands{}, Header{}, err
	}
	nextHop, commands, nextBeta, nextMac := readBeta(routingInfo)

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

	decPayload, err := AES_CTR(decKey, payload)
	if err != nil{
		return nil, err
	}

	return decPayload, nil
}