package sphinx

import "fmt"

func XorBytes(b1, b2 []byte) []byte {

	if len(b1) != len(b2){
		panic("String cannot be xored if their length is different")
	}

	b := make([]byte, len(b1))
	for i, _ := range b {
		b[i] = b1[i] ^ b2[i]
	}
	return b
}

func BytesToString(b []byte) string {
	result := ""
	for _, v := range b{
		s := fmt.Sprintf("%v", v)
		result = result + s
	}
	return result
}

