package new_packet_format

import (
	"testing"
	"crypto/elliptic"
	"math/big"
	"os"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"fmt"
	"crypto/aes"
	"anonymous-messaging/publics"
)

var curve elliptic.Curve

func TestMain(m *testing.M) {
	curve = elliptic.P224()

	os.Exit(m.Run())
}

func TestExpoSingleValue(t *testing.T) {
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil {
		t.Error(err)
	}

	randomPoint := &publics.PublicKey{Curve : curve, X : x, Y : y}
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := expo(*randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(randomPoint.X, randomPoint.Y, nBig.Bytes())
	assert.Equal(t, publics.PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)

}

func TestExpoMultipleValue(t *testing.T) {
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}
	randomPoint := &publics.PublicKey{Curve : curve, X : x, Y : y}

	var exp []big.Int
	for i := 1; i <= 5; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}

	result := expo(*randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(randomPoint.X, randomPoint.Y, big.NewInt(120).Bytes())
	assert.Equal(t, publics.PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)
}

func TestExpoBaseSingleValue(t *testing.T) {
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := expo_group_base(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(nBig.Bytes())

	assert.Equal(t, publics.PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)
}

func TestExpoBaseMultipleValue(t *testing.T){
	var exp []big.Int
	for i := 1; i <= 3; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}
	result := expo_group_base(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(big.NewInt(6).Bytes())
	assert.Equal(t, publics.PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)

}

func TestHash(t *testing.T){
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}

	randomPoint := &publics.PublicKey{Curve : curve, X : x, Y : y}
	hVal := hash(randomPoint.Bytes())

	assert.Equal(t, 32, len(hVal))

}

func TestBytesToBigNum(t *testing.T){
	bytes := big.NewInt(100).Bytes()
	result := *bytesToBigNum(curve, bytes)
	assert.Equal(t, *big.NewInt(100), result)
}

func TestGetAESKey(t *testing.T) {
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}

	randomPoint := &publics.PublicKey{Curve : curve, X : x, Y : y}
	aesKey := KDF(randomPoint.Bytes())
	assert.Equal(t, aes.BlockSize, len(aesKey))

}

func TestComputeBlindingFactor(t *testing.T){
	generator := publics.PublicKey{Curve : curve, X : curve.Params().Gx, Y : curve.Params().Gy}

	key := hash(generator.Bytes())
	b := computeBlindingFactor(curve, key)

	expected := new(big.Int)
	expected.SetString("252286146058081748716688845275111486959", 10)

	assert.Equal(t, expected, b)
}

func TestGetSharedSecrets(t *testing.T){

	pub1, _ := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()
	pub3, _ := publics.GenerateKeyPair()
	pubs := []publics.PublicKey{pub1, pub2, pub3}

	x := big.NewInt(100)


	result := getSharedSecrets(curve, pubs, *x)

	var expected []HeaderInitials
	blindFactors := []big.Int{*x}
	g := publics.PublicKey{Curve: curve, X: curve.Params().Gx, Y : curve.Params().Gy}

	v := x
	alpha0X, alpha0Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha0 := publics.PublicKey{Curve: curve, X: alpha0X, Y : alpha0Y}
	s0 := expo(pubs[0], blindFactors)
	aesS0 := KDF(s0.Bytes())
	b0:= computeBlindingFactor(curve, aesS0)

	expected = append(expected, HeaderInitials{Alpha:alpha0, Secret: s0, Blinder: *b0, SecretHash: aesS0})
	blindFactors = append(blindFactors, *b0)


	v = big.NewInt(0).Mul(v, b0)
	alpha1X, alpha1Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha1 := publics.PublicKey{Curve: curve, X: alpha1X, Y : alpha1Y}
	s1 := expo(pubs[1], blindFactors)
	aesS1 := KDF(s1.Bytes())
	b1:= computeBlindingFactor(curve, aesS1)

	expected = append(expected, HeaderInitials{Alpha:alpha1, Secret: s1, Blinder: *b1, SecretHash: aesS1})
	blindFactors = append(blindFactors, *b1)


	v = big.NewInt(0).Mul(v, b1)
	alpha2X, alpha2Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha2 := publics.PublicKey{Curve: curve, X: alpha2X, Y : alpha2Y}
	s2 := expo(pubs[2], blindFactors)
	aesS2 := KDF(s2.Bytes())
	b2:= computeBlindingFactor(curve, aesS2)

	expected = append(expected, HeaderInitials{Alpha:alpha2, Secret: s2, Blinder: *b2, SecretHash: aesS2})
	blindFactors = append(blindFactors, *b2)

	assert.Equal(t, expected, result)
}


func TestComputeFillers(t *testing.T){

	g := publics.PublicKey{Curve: curve, X: curve.Params().Gx, Y : curve.Params().Gy}
	h1 := HeaderInitials{Alpha: publics.PublicKey{}, Secret: g, Blinder: big.Int{}, SecretHash: []byte("1111111111111111")}
	h2 := HeaderInitials{Alpha: publics.PublicKey{}, Secret: g, Blinder: big.Int{}, SecretHash: []byte("1111111111111111")}
	h3 := HeaderInitials{Alpha: publics.PublicKey{}, Secret: g, Blinder: big.Int{}, SecretHash: []byte("1111111111111111")}
	tuples := []HeaderInitials{h1, h2, h3}

	pub1, _ := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()
	pub3, _ := publics.GenerateKeyPair()

	fillers := computeFillers([]publics.PublicKey{pub1,pub2,pub3}, tuples)
	fmt.Println("FILLER: ", fillers)

}

func TestXorBytesPass(t *testing.T){
	result := XorBytes([]byte("00101"), []byte("10110"))
	assert.Equal(t, []byte{1,0,0,1,1}, result)
}

func TestXorBytesFail(t *testing.T){
	result := XorBytes([]byte("00101"), []byte("10110"))
	assert.NotEqual(t, []byte("00000"), result)
}

func TestEncapsulateHeader(t *testing.T){

	pub1, _ := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()
	pub3, _ := publics.GenerateKeyPair()

	c1 := Commands{Delay: 0.34, Flag: "0"}
	c2 := Commands{Delay: 0.25, Flag: "1"}
	c3 := Commands{Delay: 1.10, Flag: "1"}
	commands := []Commands{c1, c2, c3}

	x := big.NewInt(100)
	sharedSecrets := getSharedSecrets(curve, []publics.PublicKey{pub1, pub2, pub3}, *x)

	nodesPubs := []publics.MixPubs{publics.NewMixPubs("Node1", "localhost", "3331", pub1),
									publics.NewMixPubs("Node2", "localhost", "3332", pub2),
									publics.NewMixPubs("Node3", "localhost", "3333", pub3)}

	actualHeader := encapsulateHeader(sharedSecrets, nodesPubs, []publics.PublicKey{pub1, pub2, pub3}, commands, []string{"DestinationId", "DestinationAddress", "DestKey"})

	var expectedRouting RoutingInfo
	var expectedHeader Header

	routing1 := RoutingInfo{NextHop: Hop{"DestinationId", "DestinationAddress", []byte{}}, RoutingCommands: c3,
							NextHopMetaData: []byte{}, Mac: []byte{}}

	enc_routing1 := AES_CTR(KDF(sharedSecrets[2].SecretHash), routing1.Bytes())
	mac1 := computeMac(KDF(sharedSecrets[2].SecretHash) , enc_routing1)

	routing2 := RoutingInfo{NextHop: Hop{"Node3", "localhost:3333", pub3.Bytes()}, RoutingCommands : c2,
							NextHopMetaData: enc_routing1, Mac: mac1}

	enc_routing2 := AES_CTR(KDF(sharedSecrets[1].SecretHash), routing2.Bytes())
	mac2 := computeMac(KDF(sharedSecrets[1].SecretHash) , enc_routing2)

	expectedRouting = RoutingInfo{NextHop: Hop{"Node2", "localhost:3332", pub2.Bytes()}, RoutingCommands: c1,
									NextHopMetaData: enc_routing2, Mac: mac2}

	enc_expectedRouting := AES_CTR(KDF(sharedSecrets[0].SecretHash), expectedRouting.Bytes())
	mac3 := computeMac(KDF(sharedSecrets[0].SecretHash) , enc_expectedRouting)

	expectedHeader = Header{sharedSecrets[0].Alpha, enc_expectedRouting, mac3}

	assert.Equal(t, expectedHeader, actualHeader)


}

func TestProcessSphinxPacket(t *testing.T) {

	pub1, priv1 := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()
	pub3, _ := publics.GenerateKeyPair()

	c1 := Commands{Delay: 0.34, Flag: "0"}
	c2 := Commands{Delay: 0.25, Flag: "1"}
	c3 := Commands{Delay: 1.10, Flag: "1"}
	commands := []Commands{c1, c2, c3}

	x := big.NewInt(100)
	sharedSecrets := getSharedSecrets(curve, []publics.PublicKey{pub1, pub2, pub3}, *x)

	nodesPubs := []publics.MixPubs{publics.NewMixPubs("Node1", "localhost", "3331", pub1),
		publics.NewMixPubs("Node2", "localhost", "3332", pub2),
		publics.NewMixPubs("Node3", "localhost", "3333", pub3)}

	header := encapsulateHeader(sharedSecrets, nodesPubs, []publics.PublicKey{pub1, pub2, pub3}, commands, []string{"DestinationId", "DestinationAddress", "DestKey"})

	nextHop, newCommands, newHeader := processSphinxPacket(header, priv1)

	fmt.Println(nextHop)
	fmt.Println(newCommands)
	fmt.Println(newHeader)

	assert.Equal(t, newCommands, c1)
	assert.Equal(t, nextHop.Id, "Node2")
	assert.Equal(t, nextHop.Address, "localhost:3332")
	assert.Equal(t, nextHop.PubKey, pub2.Bytes())
}
