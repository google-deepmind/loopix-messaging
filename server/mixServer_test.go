package server

import (
	"os"
	"testing"
	"anonymous-messaging/publics"
	"github.com/jmoiron/sqlx"
	"crypto/elliptic"
	"github.com/stretchr/testify/assert"
)

var mixServer MixServer

func Clean() {

	if _, err := os.Stat("testDatabase.db"); err == nil {
		errRemove := os.Remove("testDatabase.db")
		if errRemove != nil {
			panic(err)
		}
	}
}

func TestMain(m *testing.M) {
	pubM, privM := publics.GenerateKeyPair()
	mixServer = *NewMixServer("MixServer", "localhost", "9998", pubM, privM, "testDatabase.db")

	os.Exit(m.Run())
	Clean()
}

func TestSaveInPKI(t *testing.T) {
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
			Port: string(results["Port"].([]byte)), PubKey: publics.PubKeyFromBytes(elliptic.P224(), results["PubKey"].([]byte))}
		output = append(output, mix)
	}

	assert.Equal(t, 1, len(output), "There should be only one mix in the test database")

}
