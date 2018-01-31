/*
	Package helpers implements all useful functions which are used in the code of anonymous messaging system.
*/

package helpers

import (
	"math/rand"
	"time"

	"anonymous-messaging/config"
	"net"
	"anonymous-messaging/pki"
	"os"
)

func Permute(slice []config.MixPubs) []config.MixPubs {
	rand.Seed(time.Now().UTC().UnixNano())
	permutedData := make([]config.MixPubs, len(slice))
	permutation := rand.Perm(len(slice))
	for i, v := range permutation {
		permutedData[v] = slice[i]
	}
	return permutedData
}

func RandomSample(slice []config.MixPubs, length int) []config.MixPubs {
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

func AddToDatabase(pkiPath string, tableName, id, typ string, config []byte ) error {
	db, err := pki.OpenDatabase(pkiPath, "sqlite3")
	if err != nil {
		return err
	}
	defer db.Close()

	err = pki.InsertIntoTable(db, tableName, id, typ, config)
	if err != nil {
		return err
	}
	return nil
}

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}