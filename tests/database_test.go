package tests


import (
	"testing"
	"os"
	"anonymous-messaging/pki"
	"github.com/stretchr/testify/assert"
	"database/sql"
	"fmt"
)

func TestCreateDatabase(t *testing.T) {
	pki.CreateAndOpenDatabase("./testDatabase.db",  "./testDatabase.db", "sqlite3")
	_, err := os.Stat("testDatabase.db")
	assert.False(t, os.IsNotExist(err), "The database file does not exist")
}

func TestCreateTable(t *testing.T) {
	db := pki.CreateAndOpenDatabase("./testDatabase.db",  "./testDatabase.db", "sqlite3")
	params := map[string]string{"Column1" : "TEXT", "Column2" : "NUM", "Column3" : "BOOL", "Column4" : "BLOB"}

	pki.CreateTable(db, "TestTable", params)

	var exists bool
	err := db.QueryRow("SELECT CASE WHEN exists (SELECT * from sqlite_master WHERE type = 'table' AND name = 'TestTable') then 1 else 0 end").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		panic(fmt.Sprintf("Error while checking if table exists %v \n", err))
	}
	assert.True(t, exists, "Table was not added to the database")
}

func TestInsertToTable(t *testing.T) {
	db := pki.CreateAndOpenDatabase("./testDatabase.db",  "./testDatabase.db", "sqlite3")
	data := map[string]interface{}{"Column1" : "SpecialValue", "Column2" : 23, "Column3" : true, "Column4":"XYZ"}
	pki.InsertToTable(db, "TestTable", data)

	var exists bool
	err := db.QueryRow("SELECT exists (SELECT * FROM TestTable WHERE Column1=$1 AND Column2=$2 AND Column3=$3 AND Column4=$4)", "SpecialValue", 23, true, "XYZ").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("error checking if row exists %v", err)
	}
	assert.True(t, exists, "Row was not added to the database")
}

func TestQueryDatabase(t *testing.T) {
	db := pki.CreateAndOpenDatabase("./testDatabase.db",  "./testDatabase.db", "sqlite3")
	rows := pki.QueryDatabase(db, "TestTable")

	for rows.Next() {
		results := make(map[string]interface{})
		err := rows.MapScan(results)

		if err != nil {
			fmt.Printf("Error %v \n", err)
		}
		fmt.Println(results)
	}
}