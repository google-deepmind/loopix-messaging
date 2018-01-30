package server

import (
	"os"
	"testing"
	"anonymous-messaging/config"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"anonymous-messaging/sphinx"
)

var mixServer *MixServer
var providerServer *ProviderServer

const (
	TEST_DATABASE = "testDatabase.db"
)

func Clean() {

	if _, err := os.Stat(TEST_DATABASE); err == nil {
		errRemove := os.Remove(TEST_DATABASE)
		if errRemove != nil {
			panic(err)
		}
	}
}

func TestMain(m *testing.M) {
	var err error

	pubM, privM, _ := sphinx.GenerateKeyPair()
	mixServer, err = NewMixServer("MixServer", "localhost", "9998", pubM, privM, TEST_DATABASE)
	if err != nil{
		panic(m)
	}

	pubP, privP, _ := sphinx.GenerateKeyPair()
	providerServer, err = NewProviderServer("Provider", "localhost", "9997", pubP, privP, TEST_DATABASE)
	if err != nil{
		panic(m)
	}

	code := m.Run()
	os.Exit(code)
	Clean()
}

func TestMixServer_SaveInPKI(t *testing.T) {
	Clean()
	// mixServer.SaveInPKI(TEST_DATABASE)

	db, err := sqlx.Connect("sqlite3", TEST_DATABASE)
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Queryx("SELECT * FROM Mixes")
	if err != nil{
		t.Error(err)
	}
	for rows.Next() {
		result := make(map[string]interface{})
		err := rows.MapScan(result)

		if err != nil{
			t.Error(err)
		}

		pubs, err := config.MixPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, "MixServer", string(result["Id"].([]byte)), "The client id does not match")
		assert.Equal(t, "Mix", string(result["Typ"].([]byte)), "The host does not match")
		assert.Equal(t, mixServer.Config, pubs, "The config does not match")
	}
}

func TestProvider_ServerStoreMessage(t *testing.T) {
	inboxId := "ClientXYZ"
	fileName := "test.txt"
	message := []byte("Hello world message")
	providerServer.StoreMessage(message, inboxId, fileName)

	file := "./inboxes/" + inboxId + "/" + fileName
	_, err := os.Stat(file)
	assert.Nil(t, err, "The file with the message should be created")

	dat, err := ioutil.ReadFile(file)

	assert.Equal(t, message, dat, "Messages should be the same")

}

func TestProviderServer_FetchMessages(t *testing.T) {
	inboxId := "ClientXYZ"

	err := providerServer.FetchMessages(inboxId)
	if err != nil {
		t.Error(err)
	}
}