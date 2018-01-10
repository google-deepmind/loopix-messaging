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
	"strings"
)


var K = 16
var R = 5

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

type Header struct {
	Alpha PublicKey
	Beta string
	Mac string
}

func createForwardMessage(curve elliptic.Curve, pubs []PublicKey, message string) {
	createHeader(curve, pubs)
}


func createHeader(curve elliptic.Curve, pubs []PublicKey) {

	x := randomBigInt(curve.Params())

	tuples := computeSharedSecrets(curve, pubs, x)

	fillers := computeFillers(tuples)

	computeMixHeaders("Destination", "Initial", tuples, fillers)

}


func computeFillers(tuples []HeaderInitials) []string{

	fillers := []string{""}
	secrets := extractSecrets(tuples)

	for i := 1; i < len(secrets); i++ {
		f := fillers[i-1] + strings.Repeat("0", 2*K)
		fmt.Println("FILLER LENGTH: ", len(f))
		fillers = append(fillers, f)

		sHash := getAESkey(secrets[i-1])
		fmt.Println("HASH LENGTH: ", len(sHash))
	}

	return fillers

}

func extractSecrets(tuples []HeaderInitials) []PublicKey{

	var secrets []PublicKey
	for _, v := range tuples {
		secrets = append(secrets, v.Secret)
	}
	return secrets
}



func computeSharedSecrets(curve elliptic.Curve, pubs []PublicKey, initialVal big.Int) []HeaderInitials{

	blindFactors := []big.Int{initialVal}

	var tuples []HeaderInitials

	for _, mix := range pubs {

		alpha := expo_base(curve, blindFactors)

		s := expo(mix, blindFactors)
		aes_s := getAESkey(*s)

		blinder := computeBlindingFactor(aes_s)
		blindFactors = append(blindFactors, *blinder)

		hi := HeaderInitials{Alpha:*alpha, Secret: *s, Blinder: *blinder}
		tuples = append(tuples, hi)

		// TO DO ADD XORING
	}

	return tuples

}


func computeMixHeaders(destination, initial string, tuples []HeaderInitials, fillers []string){
	var headers []Header
	fmt.Println(headers)

	secrets := extractSecrets(tuples)

	beta := destination + initial + strings.Repeat("0", len(secrets))
	fmt.Println(len(beta))

	sHash := getAESkey(secrets[len(secrets) - 1])
	fmt.Println(sHash)
	fmt.Println(len(sHash))

	// TAKE CARE OF THAT v := (2*(R-len(secrets)) + 3)*K -1

	beta = xorTwoStrings(beta, string(sHash)) + fillers[len(fillers) - 1]
	fmt.Println(beta)

	


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

func xorTwoStrings(s1, s2 string) string {

	if len(s1) != len(s2){
		panic("String cannot be xored if their length is different")
	}
	b1 := []byte(s1)
	b2 := []byte(s2)

	b := make([]byte, len(s1))
	for i, _ := range b {
		b[i] = b1[i] ^ b2[i]
	}

	result := ""
	for _, v := range b{
		s := fmt.Sprintf("%v", v)
		result = result + s
	}
	return result
}