package client

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"anonymous-messaging/publics"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/sphinx"
)

var client Client
var mixPubs []publics.MixPubs
var clientPubs []publics.MixPubs
var testPacket sphinx.SphinxPacket


func clean() {
	err := os.Remove("testDatabase.db")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {

	pubC, privC := sphinx.GenerateKeyPair()
	client = *NewClient("Client", "localhost", "3332", pubC, privC, "testDatabase.db")

	code := m.Run()
	os.Exit(code)
	clean()

}

func TestClientProcessPacket(t *testing.T) {
}


func TestClientReadInMixnetPKI(t *testing.T) {

	clean()
	db, err := sqlx.Connect("sqlite3", "testDatabase.db")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		pub, _ := sphinx.GenerateKeyPair()
		mix := publics.NewMixPubs(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub)
		mixPubs = append(mixPubs, mix)
	}

	statement, e := db.Prepare("CREATE TABLE IF NOT EXISTS Mixes ( id INTEGER PRIMARY KEY, MixId TEXT, Host TEXT, Port TEXT, PubKey BLOB)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range mixPubs {
		_, err := db.Exec("INSERT INTO Mixes (MixId, Host, Port, PubKey) VALUES (?, ?, ?, ?)", elem.Id, elem.Host, elem.Port, elem.PubKey)
		if err != nil{
			panic(err)
		}
	}
	defer db.Close()

	client.ReadInMixnetPKI("testDatabase.db")

	assert.Equal(t, len(mixPubs), len(client.ActiveMixes))
	assert.Equal(t, mixPubs, client.ActiveMixes)

}

func TestClientReadInClientsPKI(t *testing.T) {

	clean()
	db, err := sqlx.Connect("sqlite3", "testDatabase.db")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		pub, _ := sphinx.GenerateKeyPair()
		client := publics.NewMixPubs(fmt.Sprintf("Client%d", i), "localhost", strconv.Itoa(3320+i), pub)
		clientPubs = append(clientPubs, client)
	}

	statement, e := db.Prepare("CREATE TABLE IF NOT EXISTS Clients ( id INTEGER PRIMARY KEY, ClientId TEXT, Host TEXT, Port TEXT, PubKey BLOB)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range clientPubs {
		db.Exec("INSERT INTO Clients (ClientId, Host, Port, PubKey) VALUES (?, ?, ?, ?)", elem.Id, elem.Host, elem.Port, elem.PubKey)
	}

	defer db.Close()


	client.ReadInClientsPKI("testDatabase.db")

	assert.Equal(t, len(clientPubs), len(client.OtherClients))
	assert.Equal(t, clientPubs, client.OtherClients)
}

func TestClientSaveInPKI(t *testing.T) {

	clean()
	SaveInPKI(client, "testDatabase.db")

	db, err := sqlx.Connect("sqlite3", "testDatabase.db")
	defer db.Close()
	if err != nil {
		panic(err)
	}

	rows, err := db.Queryx("SELECT * FROM Clients WHERE ClientId = 'Client'")
	if err != nil {
		t.Error(err)
	}

	for rows.Next() {
		results := make(map[string]interface{})
		e := rows.MapScan(results)
		if e != nil {
			t.Error(e)
		}

		assert.Equal(t, "Client", string(results["ClientId"].([]byte)), "The client id does not match")
		assert.Equal(t, "localhost", string(results["Host"].([]byte)), "The host does not match")
		assert.Equal(t, "3332", string(results["Port"].([]byte)), "The port does not match")
		assert.Equal(t, client.PubKey, results["PubKey"].([]byte), "The public key does not match")
	}
}
