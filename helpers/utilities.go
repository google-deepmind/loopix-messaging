package helpers

import (
	"anonymous-messaging/publics"
	"time"
	"math/rand"
)

func Permute(slice []publics.MixPubs) []publics.MixPubs{
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
