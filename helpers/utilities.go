/*
	Package helpers implements all useful functions which are used in the code of anonymous messaging system.
*/

package helpers

import (
	"math/rand"
	"time"

	"anonymous-messaging/publics"
	"net"
)

func Permute(slice []publics.MixPubs) []publics.MixPubs {
	rand.Seed(time.Now().UTC().UnixNano())
	permutedData := make([]publics.MixPubs, len(slice))
	permutation := rand.Perm(len(slice))
	for i, v := range permutation {
		permutedData[v] = slice[i]
	}
	return permutedData
}

func RandomSample(slice []publics.MixPubs, length int) []publics.MixPubs {
	permuted := Permute(slice)
	return permuted[:length]
}

func RandomExponential(expParam float64) float64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.ExpFloat64() / expParam
}

func ResolveTCPAddress(host, port string) (*net.TCPAddr, error) {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}
	return addr, nil
}