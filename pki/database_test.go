package pki


import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
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
	CreateAndOpenDatabase("./testDatabase.db",  "./testDatabase.db", "sqlite3")
	_, err := os.Stat("testDatabase.db")
	assert.False(t, os.IsNotExist(err), "The database file does not exist")
}

func TestCreateTable(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", "./testDatabase2.db")

	if err != nil {
		panic(err)
	}

	params := map[string]string{"Column1" : "TEXT", "Column2" : "NUM", "Column3" : "BIT", "Column4" : "BLOB"}
	CreateTable(db, "TestTable", params)

	var exists bool
	err = db.QueryRow("SELECT CASE WHEN exists (SELECT * from sqlite_master WHERE type = 'table' AND name = 'TestTable') then 1 else 0 end").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		panic(fmt.Sprintf("Error while checking if table exists %v \n", err))
	}
	assert.True(t, exists, "Table was not added to the database")
}

func TestInsertToTable(t *testing.T) {
	db, e := sqlx.Connect("sqlite3", "./testDatabase2.db")

	if e != nil {
		panic(e)
	}


	data := map[string]interface{}{"Column1" : "SpecialValue", "Column2" : 23, "Column3" : true, "Column4":"XYZ"}
	InsertToTable(db, "TestTable", data)

	var exists bool
	err := db.QueryRow("SELECT exists (SELECT * FROM TestTable WHERE Column1=$1 AND Column2=$2 AND Column3=$3 AND Column4=$4)", "SpecialValue", 23, true, "XYZ").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("error checking if row exists %v", err)
	}
	assert.True(t, exists, "Row was not added to the database")
}

func TestQueryDatabase(t *testing.T) {
	db := CreateAndOpenDatabase("./testDatabase2.db",  "./testDatabase2.db", "sqlite3")

	rows := QueryDatabase(db, "TestTable")

 	for rows.Next() {
		results := make(map[string]interface{})
		err := rows.MapScan(results)

		if err != nil {
			fmt.Printf("Error %v \n", err)
		}

		assert.Equal(t, "SpecialValue", string(results["Column1"].([]byte)), "Values should be equal")
		assert.Equal(t, int64(23), results["Column2"].(int64), "Values should be equal")
		assert.Equal(t, int64(1), results["Column3"].(int64), "Values should be equal")
		assert.Equal(t, "XYZ", string(results["Column4"].([]byte)), "Values should be equal")
	}
}