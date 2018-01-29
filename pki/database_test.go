package pki

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"database/sql"
)

func Setup() {

	f, err := os.Create("./testDatabase2.db")
	if err != nil {
		panic(err)
	}

	defer f.Close()
}

func Clean() {

	if _, err := os.Stat("./testDatabase.db"); err == nil {
		errRemove := os.Remove("./testDatabase.db")
		if errRemove != nil {
			panic(err)
		}
	}

	if _, err := os.Stat("./testDatabase2.db"); err == nil {
		errRemove := os.Remove("./testDatabase2.db")
		if errRemove != nil {
			panic(err)
		}
	}
}

func TestMain(m *testing.M) {

	Setup()
	code := m.Run()
	Clean()
	os.Exit(code)
}

func TestCreateDatabase(t *testing.T) {
	_, err := OpenDatabase("./testDatabase.db", "sqlite3")
	if err != nil{
		t.Error(err)
	}

	_, err = os.Stat("testDatabase.db")
	if err != nil{
		t.Error(err)
	}
	assert.False(t, os.IsNotExist(err), "The database file does not exist")
}

func TestCreateTable(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", "./testDatabase2.db")

	if err != nil {
		panic(err)
	}

	params := map[string]string{"Id": "TEXT", "Typ": "TEXT", "Config": "BLOB"}
	err = CreateTable(db, "TestTable", params)
	if err != nil{
		t.Error(err)
	}

	var exists bool
	err = db.QueryRow("SELECT CASE WHEN exists (SELECT * from sqlite_master WHERE type = 'table' AND name = 'TestTable') then 1 else 0 end").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		panic(fmt.Sprintf("Error while checking if table exists %v \n", err))
	}
	assert.True(t, exists, "Table was not added to the database")
}

func TestInsertToTable(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", "./testDatabase2.db")

	if err != nil {
		t.Error(err)
	}

	err = InsertIntoTable(db, "TestTable", "TestVal1", "TestVal2", []byte("TestVal3"))

	if err != nil{
		t.Error(err)
	}

	var exists bool
	err = db.QueryRow("SELECT exists (SELECT * FROM TestTable WHERE Id=$1 AND Typ=$2 AND Config=$3)", "TestVal1", "TestVal2", []byte("TestVal3")).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		t.Error(err)
	}

	assert.True(t, exists, "Row was not added to the database")
}

func TestQueryDatabase(t *testing.T) {
	db, err := OpenDatabase("./testDatabase2.db", "sqlite3")

	if err != nil{
		t.Error(err)
	}

	rows, err := QueryDatabase(db, "TestTable")

	if err != nil{
		t.Error(err)
	}

	for rows.Next() {
		results := make(map[string]interface{})
		err := rows.MapScan(results)

		if err != nil {
			fmt.Printf("Error %v \n", err)
		}

		assert.Equal(t, "TestVal1", string(results["Id"].([]byte)), "Values should be equal")
		assert.Equal(t, "TestVal2", string(results["Typ"].([]byte)), "Values should be equal")
		assert.Equal(t, []byte("TestVal3"), results["Config"].([]byte), "Values should be equal")
	}
}

func TestInsertIntoTable(t *testing.T) {
	db, err := OpenDatabase("./testDatabase2.db", "sqlite3")

	if err != nil{
		t.Error(err)
	}

	params := make(map[string]string)
	params["Id"] = "TEXT"
	params["Typ"] = "TEXT"
	params["Config"] = "BLOB"

	CreateTable(db, "InsertTestTable", params)

	err = InsertIntoTable(db, "InsertTestTable", "MyId", "MyTyp", []byte("Some bytes"))

	if err != nil {
		t.Error(err)
	}

	rows, err := db.Queryx("SELECT * FROM InsertTestTable")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		results := make(map[string]interface{})
		err := rows.MapScan(results)

		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, "MyId", string(results["Id"].([]byte)), "Should be equal")
		assert.Equal(t, "MyTyp", string(results["Typ"].([]byte)), "Should be equal")
		assert.Equal(t, []byte("Some bytes"), results["Config"].([]byte), "Should be equal")
	}

}