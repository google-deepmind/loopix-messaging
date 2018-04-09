// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
	Package helpers implements all useful functions which are used in the code of anonymous messaging system.
*/

package helpers

import (
	"anonymous-messaging/config"
	"anonymous-messaging/pki"

	"github.com/protobuf/proto"

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

// SHA256 computes the hash value of a given argument using SHA256 algorithm.
func SHA256(arg []byte) []byte {
	h := sha256.New()
	h.Write([]byte(arg))
	return h.Sum(nil)
}

func GetMixesPKI(pkiDir string) ([]config.MixConfig, error) {
	var mixes []config.MixConfig

	db, err := pki.OpenDatabase(pkiDir, "sqlite3")
	if err != nil {
		return nil, err
	}

	recordsMixes, err := pki.QueryDatabase(db, "Pki", "Mix")
	if err != nil {
		return nil, err
	}

	for recordsMixes.Next() {
		result := make(map[string]interface{})
		err := recordsMixes.MapScan(result)
		if err != nil {
			return nil, err
		}

		var mixConfig config.MixConfig
		err = proto.Unmarshal(result["Config"].([]byte), &mixConfig)
		if err != nil {
			return nil, err
		}
		mixes = append(mixes, mixConfig)
	}

	return mixes, nil
}

func GetClientPKI(pkiDir string) ([]config.ClientConfig, error) {
	var clients []config.ClientConfig

	db, err := pki.OpenDatabase(pkiDir, "sqlite3")
	if err != nil {
		return nil, err
	}

	recordsClients, err := pki.QueryDatabase(db, "Pki", "Client")
	if err != nil {
		return nil, err
	}
	for recordsClients.Next() {
		result := make(map[string]interface{})
		err := recordsClients.MapScan(result)

		if err != nil {
			return nil, err
		}

		var clientConfig config.ClientConfig
		err = proto.Unmarshal(result["Config"].([]byte), &clientConfig)
		if err != nil {
			return nil, err
		}

		clients = append(clients, clientConfig)
	}
	return clients, nil
}

func GetLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}

	return "", errors.New("Couldn't find the valid IP. Check internet connection.")
}
