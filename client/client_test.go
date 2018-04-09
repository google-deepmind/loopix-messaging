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

package client

import (
	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"

	"github.com/jmoiron/sqlx"
	"github.com/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"fmt"
	"os"
	"strconv"
	"testing"
)

var providerPubs config.MixConfig
var testPacket sphinx.SphinxPacket
var testMixSet []config.MixConfig
var testClientSet []config.ClientConfig

const (
	pkiDir = "testDatabase.db"
)

func setupTestDatabase() (*sqlx.DB, error) {

	db, err := sqlx.Connect("sqlite3", pkiDir)
	if err != nil {
		return nil, err
	}

	query := `CREATE TABLE Pki (
		idx INTEGER PRIMARY KEY,
    	Id TEXT,
    	Typ TEXT,
    	Config BLOB);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return db, err
}

func SetupTestMixesInDatabase(t *testing.T) error {
	clean()

	db, err := setupTestDatabase()
	if err != nil {
		t.Fatal(err)
	}

	insertQuery := `INSERT INTO Pki (Id, Typ, Config) VALUES (?, ?, ?)`

	for i := 0; i < 10; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil {
			return err
		}
		m := config.MixConfig{Id: fmt.Sprintf("Mix%d", i),
			Host:   "localhost",
			Port:   strconv.Itoa(9980 + i),
			PubKey: pub}
		mBytes, err := proto.Marshal(&m)
		if err != nil {
			return err
		}
		_, err = db.Exec(insertQuery, m.Id, "Mix", mBytes)
		if err != nil {
			return err
		}
		testMixSet = append(testMixSet, m)
	}
	return nil
}

func SetupTestClientsInDatabase(t *testing.T) {
	clean()

	db, err := setupTestDatabase()
	if err != nil {
		t.Fatal(err)
	}

	insertQuery := `INSERT INTO Pki (Id, Typ, Config) VALUES (?, ?, ?)`

	for i := 0; i < 10; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil {
			t.Fatal(err)
		}
		c := config.ClientConfig{Id: fmt.Sprintf("Client%d", i),
			Host:   "localhost",
			Port:   strconv.Itoa(9980 + i),
			PubKey: pub}
		cBytes, err := proto.Marshal(&c)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(insertQuery, c.Id, "Client", cBytes)
		if err != nil {
			t.Fatal(err)
		}
		testClientSet = append(testClientSet, c)
	}
}

func SetupTestClient(t *testing.T) *client {
	pubP, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	providerPubs = config.MixConfig{Id: "Provider", Host: "localhost", Port: "9995", PubKey: pubP}

	pubC, privC, err := sphinx.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewTestClient("Client", "localhost", "3332", pubC, privC, pkiDir, providerPubs)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func clean() error {
	if _, err := os.Stat(pkiDir); err == nil {
		err := os.Remove(pkiDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestMain(m *testing.M) {

	defer clean()

	code := m.Run()
	clean()
	os.Exit(code)

}

func TestClient_GetMessagesFromProvider(t *testing.T) {

}

// TODO: Fix this test
//func TestClient_RegisterToken_Pass(t *testing.T) {
//client := SetupTestClient(t)
//client.RegisterToken([]byte("TestToken"))
//r := <- client.registrationDone
//assert.True(t, r)
//assert.Equal(t, []byte("TestToken"), client.token, "Client should register only given token")
//}

//func TestClient_RegisterToken_Fail(t *testing.T) {
//	client := SetupTestClient(t)
//	client.RegisterToken([]byte("TestToken"))
//	assert.NotEqual(t, []byte("WrongToken"), client.token, "Client should register only the given token")
//}

// TODO: Fix this test
func TestClient_RegisterToProvider(t *testing.T) {

}

// TODO: Fix this test
//func TestClient_SendMessage(t *testing.T) {
//	pubP, _, err := sphinx.GenerateKeyPair()
//	if err != nil{
//		t.Fatal(err)
//	}
//	providerPubs = config.MixConfig{Id: "Provider", Host: "localhost", Port: "9995", PubKey: pubP}
//
//	pubR, _, err := sphinx.GenerateKeyPair()
//	if err != nil{
//		t.Fatal(err)
//	}
//	recipient := config.ClientConfig{Id:"Recipient", Host:"localhost", Port:"9999", PubKey: pubR, Provider: &providerPubs}
//	fmt.Println(recipient)
//	pubM1, _, err := sphinx.GenerateKeyPair()
//	if err != nil{
//		t.Fatal(err)
//	}
//	pubM2, _, err := sphinx.GenerateKeyPair()
//	if err != nil{
//		t.Fatal(err)
//	}
//	m1 := config.MixConfig{Id: "Mix1", Host: "localhost", Port: strconv.Itoa(9980), PubKey: pubM1}
//	m2 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: strconv.Itoa(9981), PubKey: pubM2}
//
//	client := SetupTestClient(t)
//	client.ActiveMixes = []config.MixConfig{m1, m2}
//
//	addr, err := helpers.ResolveTCPAddress(client.Host, client.Port)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	client.Listener, err = net.ListenTCP("tcp", addr)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = client.SendMessage("TestMessage", recipient)
//	if err != nil{
//		t.Fatal(err)
//	}
//	err = client.Listener.Close()
//	if err != nil{
//		t.Fatal(err)
//	}
//}

// TODO: Fix this test
func TestClient_ProcessPacket(t *testing.T) {

}

func TestClient_ReadInMixnetPKI(t *testing.T) {
	clean()
	SetupTestMixesInDatabase(t)

	client := SetupTestClient(t)
	err := client.ReadInNetworkFromPKI("testDatabase.db")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(testMixSet), len(client.Network.Mixes))
	assert.Equal(t, testMixSet, client.Network.Mixes)

}
