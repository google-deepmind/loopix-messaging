package new_packet_format

import (
	"testing"
	"crypto/elliptic"
	"math/big"
	"os"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
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

	randomPoint := &PublicKey{Curve : curve, X : x, Y : y}
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := *expo(*randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(randomPoint.X, randomPoint.Y, nBig.Bytes())
	assert.Equal(t, PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)

}

func TestExpoMultipleValue(t *testing.T) {
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}
	randomPoint := &PublicKey{Curve : curve, X : x, Y : y}

	var exp []big.Int
	for i := 1; i <= 5; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}

	result := *expo(*randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(randomPoint.X, randomPoint.Y, big.NewInt(120).Bytes())
	assert.Equal(t, PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)
}

func TestExpoBaseSingleValue(t *testing.T) {
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := *expo_base(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(nBig.Bytes())

	assert.Equal(t, PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)
}

func TestExpoBaseMultipleValue(t *testing.T){
	var exp []big.Int
	for i := 1; i <= 3; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}
	result := *expo_base(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(big.NewInt(6).Bytes())
	assert.Equal(t, PublicKey{Curve: curve, X: expectedX, Y: expectedY}, result)

}

func TestHash(t *testing.T){
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}

	randomPoint := &PublicKey{Curve : curve, X : x, Y : y}
	hVal := hash(*randomPoint)

	assert.Equal(t, 32, len(hVal))

}

func TestBytesToBigNum(t *testing.T){
	bytes := big.NewInt(100).Bytes()
	result := *bytesToBigNum(bytes)
	assert.Equal(t, *big.NewInt(100), result)
}

func TestGetAESkey(t *testing.T) {
	_, x, y, err  := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil{
		t.Error(err)
	}

	randomPoint := &PublicKey{Curve : curve, X : x, Y : y}
	aesKey := getAESkey(*randomPoint)
	assert.Equal(t, 32, len(aesKey))

}

func TestComputeBlindingFactor(t *testing.T){
	generator := PublicKey{Curve : curve, X : curve.Params().Gx, Y : curve.Params().Gy}

	key := hash(generator)
	b := computeBlindingFactor(key)

	expected := new(big.Int)
	expected.SetString("8411123853944643709977978256681440100", 10)

	assert.Equal(t, expected, b)
}

func TestCreateHeader(t *testing.T){

	var pubs []PublicKey

	for i := 1; i <=3; i++ {
		pointX, pointY := curve.ScalarBaseMult(big.NewInt(6).Bytes())
		point := &PublicKey{Curve : curve, X : pointX, Y : pointY}
		pubs = append(pubs, *point)
	}

	x := big.NewInt(100)


	result := computeSharedSecrets(curve, pubs, *x)

	var expected []HeaderInitials
	blindFactors := []big.Int{*x}
	g := PublicKey{Curve: curve, X: curve.Params().Gx, Y : curve.Params().Gy}

	v := x
	alpha0X, alpha0Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha0 := PublicKey{Curve: curve, X: alpha0X, Y : alpha0Y}
	s0 := expo(pubs[0], blindFactors)
	b0:= computeBlindingFactor(getAESkey(*s0))

	expected = append(expected, HeaderInitials{Alpha:alpha0, Secret: *s0, Blinder: *b0})
	blindFactors = append(blindFactors, *b0)


	v = big.NewInt(0).Mul(v, b0)
	alpha1X, alpha1Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha1 := PublicKey{Curve: curve, X: alpha1X, Y : alpha1Y}
	s1 := expo(pubs[1], blindFactors)
	b1:= computeBlindingFactor(getAESkey(*s1))

	expected = append(expected, HeaderInitials{Alpha:alpha1, Secret: *s1, Blinder: *b1})
	blindFactors = append(blindFactors, *b1)


	v = big.NewInt(0).Mul(v, b1)
	alpha2X, alpha2Y := curve.Params().ScalarMult(g.X, g.Y, v.Bytes())
	alpha2 := PublicKey{Curve: curve, X: alpha2X, Y : alpha2Y}
	s2 := expo(pubs[2], blindFactors)
	b2:= computeBlindingFactor(getAESkey(*s2))

	expected = append(expected, HeaderInitials{Alpha:alpha2, Secret: *s2, Blinder: *b2})
	blindFactors = append(blindFactors, *b2)

	assert.Equal(t, expected, result)


}


func TestComputeFillers(t *testing.T){

	g := PublicKey{Curve: curve, X: curve.Params().Gx, Y : curve.Params().Gy}
	h1 := HeaderInitials{Alpha: PublicKey{}, Secret: g, Blinder: big.Int{}}
	h2 := HeaderInitials{Alpha: PublicKey{}, Secret: g, Blinder: big.Int{}}
	h3 := HeaderInitials{Alpha: PublicKey{}, Secret: g, Blinder: big.Int{}}
	tuples := []HeaderInitials{h1, h2, h3}

	fillers := computeFillers(tuples)

	computeMixHeaders("Destination", "InitialVector11111", tuples, fillers)
}

func TestXorTwoStringsPass(t *testing.T){
	result := xorTwoStrings("00101", "10110")
	assert.Equal(t, "10011", result)
}

func TestXorTwoStringsFail(t *testing.T){
	result := xorTwoStrings("00101", "10110")
	assert.NotEqual(t, "00000", result)
}