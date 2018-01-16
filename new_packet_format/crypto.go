package new_packet_format

import (
	"crypto/sha256"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
)

func AES_CTR(key, plaintext []byte) []byte {

	ciphertext := make([]byte, len(plaintext))

	iv := []byte("0000000000000000")
	//if _, err := io.ReadFull(crand.Reader, iv); err != nil {
	//	panic(err)
	//}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext
}

func hash(arg []byte) []byte{

	h := sha256.New()
	h.Write(arg)

	return h.Sum(nil)
}

func Hmac(key, message []byte) []byte{
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}