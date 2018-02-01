package pki

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"database/sql"
	"github.com/stretchr/testify/assert"
)


const (
	TESTDATABASE = "./testDatabase.db"
)


var db *sqlx.DB

func Setup() (*sqlx.DB, error) {

	Clean()

	db, err := sqlx.Connect("sqlite3", TESTDATABASE)
	if err != nil {
		return nil, err
	}

	query := `CREATE TABLE TableXX (
		idx INTEGER PRIMARY KEY,
    	Id TEXT,
    	Typ TEXT,
    	Config BLOB);`

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	insertQuery := `INSERT INTO TableXX (Id, Typ, Config) VALUES (?, ?, ?)`

	_, err = db.Exec(insertQuery, "ABC", "DEF", []byte("GHI"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Clean() {

	if _, err := os.Stat(TESTDATABASE); err == nil {
		errRemove := os.Remove(TESTDATABASE)
		if errRemove != nil {
			panic(err)
		}
	}
}

func TestMain(m *testing.M) {

	var err error
	db, err = Setup()
	if err != nil {
		fmt.Println(err)
		panic(m)
	}
	defer db.Close()

	code := m.Run()
	Clean()
	os.Exit(code)
}

func TestCreateTable(t *testing.T) {

	params := map[string]string{"Id": "TEXT", "Typ": "TEXT", "Config": "BLOB"}
	err := CreateTable(db, "TestTable", params)
	if err != nil {
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

	err := InsertIntoTable(db, "TestTable", "TestVal1", "TestVal2", []byte("TestVal3"))
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
	rows, err := QueryDatabase(db, "TableXX")

	if err != nil{
		t.Error(err)
	}

	results := make(map[string]interface{})
	for rows.Next() {
		err := rows.MapScan(results)

		if err != nil {
			fmt.Printf("Error %v \n", err)
		}
	}
	assert.Equal(t, "ABC", string(results["Id"].([]byte)), "Should be equal")
	assert.Equal(t, "DEF", string(results["Typ"].([]byte)), "Should be equal")
	assert.Equal(t, []byte("GHI"), results["Config"].([]byte), "Should be equal")

}

func TestInsertIntoTable(t *testing.T) {

	err := InsertIntoTable(db, "TableXX", "TestInsertId", "TestInsertTyp", []byte("TestInsertBytes"))
	if err != nil {
		t.Error(err)
	}

	exists, err := rowExists(db,"SELECT * FROM TableXX WHERE Id=$1 AND Typ=$2 AND Config=$3", "TestInsertId", "TestInsertTyp", []byte("TestInsertBytes"))
	if err != nil {
		t.Error(err)
	}
	assert.True(t, exists, "The inserted row was not found in the database")
}