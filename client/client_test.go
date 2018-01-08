package client

import (
	"anonymous-messaging/packet_format"
	"anonymous-messaging/publics"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
)

var client Client
var mixPubs []publics.MixPubs
var clientPubs []publics.MixPubs
var testPacket packet_format.Packet

func makeTestInMemoryPKI() {
	db, err := sqlx.Connect("sqlite3", "testDatabase.db")
	defer db.Close()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		mix := publics.NewMixPubs(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), 0)
		mixPubs = append(mixPubs, mix)
	}

	statement, e := db.Prepare("CREATE TABLE IF NOT EXISTS Mixes ( id INTEGER PRIMARY KEY, MixId TEXT, Host TEXT, Port TEXT, PubKey NUM)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range mixPubs {
		db.Exec("INSERT INTO Mixes (MixId, Host, Port, PubKey) VALUES (?, ?, ?, ?)", elem.Id, elem.Host, elem.Port, elem.PubKey)
	}

	for i := 0; i < 10; i++ {
		client := publics.NewMixPubs(fmt.Sprintf("Client%d", i), "localhost", strconv.Itoa(3320+i), 0)
		clientPubs = append(clientPubs, client)
	}

	statement, e = db.Prepare("CREATE TABLE IF NOT EXISTS Clients ( id INTEGER PRIMARY KEY, ClientId TEXT, Host TEXT, Port TEXT, PubKey NUM)")
	if e != nil {
		panic(e)
	}
	statement.Exec()

	for _, elem := range clientPubs {
		db.Exec("INSERT INTO Clients (ClientId, Host, Port, PubKey) VALUES (?, ?, ?, ?)", elem.Id, elem.Host, elem.Port, elem.PubKey)
	}

	defer db.Close()
}

func clean() {
	err := os.Remove("testDatabase.db")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	client = *NewClient("Client", "localhost", "3332", "../pki/database.db", 0, 0)
	testPacket = packet_format.NewPacket("Message", nil, nil, nil)

	makeTestInMemoryPKI()

	code := m.Run()

	clean()
	os.Exit(code)

}

func TestClientProcessPacket(t *testing.T) {
	packet := packet_format.NewPacket("Message", []float64{0.1, 0.2, 0.3}, mixPubs, nil)

	m := client.ProcessPacket(packet)
	assert.Equal(t, m, "Message", "The final message should be the same as the init one")
}

func TestClientSendMessage(t *testing.T) {
	// TO DO
}

func TestClientSend(t *testing.T) {
	// TO DO
}

func TestClientListenForConnections(t *testing.T) {
	// TO DO
}

func TestClientHandleConnection(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go client.HandleConnection(clientConn)
	serverConn.Write([]byte(packet_format.ToString(testPacket)))

	// How I should now check that HandleConnection performed what it was suppose to do? Should I mock?
}

func TestClientStart(t *testing.T) {
	// TO DO
}

func TestClientRun(t *testing.T) {
	// TO DO
}

func TestClientReadInMixnetPKI(t *testing.T) {
	client.ReadInMixnetPKI("testDatabase.db")
	if len(mixPubs) == len(client.ActiveMixes) {

	} else {
		t.Error("Reading mixes incorrect")
	}
}

func TestClientReadInClientsPKI(t *testing.T) {
	client.ReadInClientsPKI("testDatabase.db")
	if len(clientPubs) == len(client.OtherClients) {

	} else {
		t.Error("Reading mixes incorrect")
	}
}

func TestClientSaveInPKI(t *testing.T) {
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
		assert.Equal(t, int64(0), results["PubKey"].(int64), "The public key does not match")
	}
}
