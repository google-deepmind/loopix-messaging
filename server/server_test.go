package server

import (
	"os"
	"testing"
	"anonymous-messaging/publics"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"anonymous-messaging/sphinx"
)

var mixServer MixServer
var providerServer ProviderServer

func Clean() {

	if _, err := os.Stat("testDatabase.db"); err == nil {
		errRemove := os.Remove("testDatabase.db")
		if errRemove != nil {
			panic(err)
		}
	}
}

func TestMain(m *testing.M) {
	pubM, privM := sphinx.GenerateKeyPair()
	mixServer = *NewMixServer("MixServer", "localhost", "9998", pubM, privM, "testDatabase.db")

	pubP, privP := sphinx.GenerateKeyPair()
	providerServer = *NewProviderServer("Provider", "localhost", "9997", pubP, privP, "testDatabase.db")

	code := m.Run()
	Clean()
	os.Exit(code)
}

func TestMixServer_SaveInPKI(t *testing.T) {
	Clean()
	mixServer.SaveInPKI("testDatabase.db")

	db, err := sqlx.Connect("sqlite3", "testDatabase.db")
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Queryx("SELECT * FROM Mixes")
	if err != nil{
		t.Error(err)
	}

	output := []publics.MixPubs{}
	for rows.Next() {
		results := make(map[string]interface{})
		err := rows.MapScan(results)

		if err != nil{
			t.Error(err)
		}
		mix := publics.MixPubs{Id: string(results["MixId"].([]byte)), Host: string(results["Host"].([]byte)),
			Port: string(results["Port"].([]byte)), PubKey: results["PubKey"].([]byte)}
		output = append(output, mix)
	}

	assert.Equal(t, 1, len(output), "There should be only one mix in the test database")

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