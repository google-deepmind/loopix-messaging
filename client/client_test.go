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
	"anonymous-messaging/server"
)

var client Client
var providerPubs publics.MixPubs
var testPacket sphinx.SphinxPacket


func clean() {
	err := os.Remove("testDatabase.db")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {

	pubP, _ := sphinx.GenerateKeyPair()
	providerPubs = publics.MixPubs{Id: "Provider", Host: "localhost", Port: "9995", PubKey: pubP}

	pubC, privC := sphinx.GenerateKeyPair()
	client = *NewClient("Client", "localhost", "3332", pubC, privC, "testDatabase.db", providerPubs)

	code := m.Run()
	clean()
	os.Exit(code)

}

func TestClient_ProcessPacket(t *testing.T) {
}


func TestClient_ReadInMixnetPKI(t *testing.T) {

	clean()
	db, err := sqlx.Connect("sqlite3", "testDatabase.db")

	if err != nil {
		panic(err)
	}

	// TO DO: fix this test
	var mixes []server.MixServer
	var mixPubs []publics.MixPubs
	for i := 0; i < 10; i++ {
		pub, priv := sphinx.GenerateKeyPair()
		mix := server.MixServer{fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub, priv, "testDatabase.db"}
		mixes = append(mixes, mix)
		mixPubs = append(mixPubs, mix.Config)
	}

	statement, e := db.Prepare("CREATE TABLE IF NOT EXISTS Mixes ( id INTEGER PRIMARY KEY, Id TEXT, Typ TEXT, Config BLOB)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range mixes {
		_, err := db.Exec("INSERT INTO Mixes (Id, Typ, Config) VALUES (?, ?, ?)", elem.Id, "Mix", elem.Config)
		if err != nil{
			panic(err)
		}
	}
	defer db.Close()

	client.ReadInMixnetPKI("testDatabase.db")

	assert.Equal(t, len(mixes), len(client.ActiveMixes))
	assert.Equal(t, mixPubs, client.ActiveMixes)

}

func TestClient_ReadInClientsPKI(t *testing.T) {

	clean()
	db, err := sqlx.Connect("sqlite3", "testDatabase.db")

	if err != nil {
		panic(err)
	}

	var clientsList []Client
	var clientsPubs []publics.ClientPubs
	for i := 0; i < 5; i++ {
		pub, priv := sphinx.GenerateKeyPair()
		client := NewClient(fmt.Sprintf("Client%d", i), "localhost", strconv.Itoa(3320+i), pub, priv, "testDatabase.db", providerPubs)
		clientsList = append(clientsList, *client)
		clientsPubs = append(clientsPubs, client.Config)
	}

	statement, e := db.Prepare("CREATE TABLE IF NOT EXISTS Clients ( id INTEGER PRIMARY KEY, Id TEXT, Typ TEXT, Config BLOB)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range clientsList {
		db.Exec("INSERT INTO Clients (Id, Typ, Config) VALUES (?, ?, ?)", elem.Id, "Client", elem.Config)
	}

	defer db.Close()


	client.ReadInClientsPKI("testDatabase.db")

	assert.Equal(t, len(clientsList), len(client.OtherClients))
	assert.Equal(t, clientsPubs, client.OtherClients)
}

func TestClient_SaveInPKI(t *testing.T) {

	clean()
	SaveInPKI(client, "testDatabase.db")

	db, err := sqlx.Connect("sqlite3", "testDatabase.db")
	defer db.Close()
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Queryx("SELECT * FROM Clients WHERE Id = 'Client'")
	if err != nil {
		t.Error(err)
	}

	for rows.Next() {
		result := make(map[string]interface{})
		err = rows.MapScan(result)
		if err != nil {
			t.Error(err)
		}

		pubs, err := publics.ClientPubsFromBytes(result["Config"].([]byte))
		if err != nil {
			t.Error(err)
		}


		assert.Equal(t, "Client", string(result["Id"].([]byte)), "The client id does not match")
		assert.Equal(t, "Client", string(result["Typ"].([]byte)), "The host does not match")
		assert.Equal(t, client.Config, pubs, "The config does not match")
	}
}
