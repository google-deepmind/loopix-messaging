package new_packet_format

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"crypto/sha256"
)

type NewPacketFormat struct {

}

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

type PrivateKey struct {
	privk []byte
}

type HeaderInitials struct {
	Alpha PublicKey
	Secret PublicKey
	Blinder big.Int
}

func createHeader(curve elliptic.Curve, pubs []PublicKey) {

	x := randomBigInt(curve.Params())

	computeSharedSecrets(curve, pubs, x)

}

func computeSharedSecrets(curve elliptic.Curve, pubs []PublicKey, initialVal big.Int) []HeaderInitials{

	blindFactors := []big.Int{initialVal}

	var tuples []HeaderInitials

	for _, mix := range pubs {

		alpha := expo_base(curve, blindFactors)
		Nothing(alpha, &mix)

		s := expo(mix, blindFactors)
		aes_s := getAESkey(*s)

		blinder := computeBlindingFactor(aes_s)
		blindFactors = append(blindFactors, *blinder)

		hi := HeaderInitials{Alpha:*alpha, Secret: *s, Blinder: *blinder}
		tuples = append(tuples, hi)
	}

	return tuples

}

func Nothing(p *PublicKey, key *PublicKey) {

}

func computeBlindingFactor(key []byte) *big.Int{
	iv := []byte("initialvector000")
	blinderBytes := computeSharedSecretHash(key, iv)
	blinder := bytesToBigNum(blinderBytes)

	return blinder
}

func computeSharedSecretHash(key []byte, iv []byte) []byte{
	aesCipher, err := aes.NewCipher(key)

	if err != nil {
		panic(err)
	}

	cbc := cipher.NewCBCEncrypter(aesCipher, iv)
	plaintext := []byte("0000000000000000")

	ciphertext := make([]byte, len(plaintext))
	cbc.CryptBlocks(ciphertext, plaintext)

	return ciphertext
}

func expo(base PublicKey, exp []big.Int) *PublicKey{
	x := exp[0]
	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}
	curve := base.Curve
	resultX, resultY := curve.Params().ScalarMult(base.X, base.Y, x.Bytes())
	result := PublicKey{curve, resultX, resultY}
	return &result
}

func expo_base(curve elliptic.Curve, exp []big.Int) *PublicKey{
	x := exp[0]

	for _, val := range exp[1:] {
		x = *big.NewInt(0).Mul(&x, &val)
	}

	resultX, resultY := curve.Params().ScalarBaseMult(x.Bytes())
	return &PublicKey{Curve: curve, X: resultX, Y: resultY}

}

func getAESkey(p PublicKey) []byte{
	return hash(p)
}

func hash(k PublicKey) []byte{

	structAsString, err := json.Marshal(&k)

	if err != nil{
		panic(err)
	}

	bytes := []byte(structAsString)

	h := sha256.New()
	h.Write(bytes)

	return h.Sum(nil)
}


func bytesToBigNum(value []byte) *big.Int{
	nBig := new(big.Int)
	nBig.SetBytes(value)

	return nBig
}

func sumSlice(s []int) int{
	sum := 0
	for _, value := range s {
		sum += value
	}
	return sum
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