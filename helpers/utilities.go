/*
	Package helpers implements all useful functions which are used in the code of anonymous messaging system.
*/

package helpers

import (
	"anonymous-messaging/config"
	"anonymous-messaging/pki"

	"crypto/sha256"
	"errors"
	"math/rand"
	"net"
	"os"
	"time"
)

func Permute(slice []config.MixConfig) ([]config.MixConfig, error) {
	if len(slice) == 0 {
		return nil, errors.New(" cannot permute an empty list of mixes")
	}

	rand.Seed(time.Now().UTC().UnixNano())
	permutedData := make([]config.MixConfig, len(slice))
	permutation := rand.Perm(len(slice))
	for i, v := range permutation {
		permutedData[v] = slice[i]
	}
	return permutedData, nil
}

func RandomSample(slice []config.MixConfig, length int) ([]config.MixConfig, error) {
	if len(slice) < length {
		return nil, errors.New(" cannot take a sample larger than the given list")
	}

	permuted, err := Permute(slice)
	if err != nil {
		return nil, err
	}

	return permuted[:length], err
}

func RandomExponential(expParam float64) (float64, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	if expParam <= 0.0 {
		return 0.0, errors.New("the parameter of exponential distribution has to be larger than zero")
	}
	return rand.ExpFloat64() / expParam, nil
}

func ResolveTCPAddress(host, port string) (*net.TCPAddr, error) {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// TO DO: This function is useless; remove it and change the code

func AddToDatabase(pkiPath string, tableName, id, typ string, config []byte) error {
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
	if os.IsNotExist(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

/*
	SHA256 computes the hash value of a given argument using SHA256 algorithm.
*/
func SHA256(arg []byte) []byte {
	h := sha256.New()
	h.Write([]byte(arg))
	return h.Sum(nil)
}
