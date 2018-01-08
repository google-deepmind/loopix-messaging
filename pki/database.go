package pki

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabase(dataSourceName, dbDriver string) *sqlx.DB {

	var db *sqlx.DB
	db, err := sqlx.Connect(dbDriver, dataSourceName)

	if err != nil {
		panic(err)
	}

	return db
}

func CreateTable(db *sqlx.DB, tableName string, params map[string]string) {
	paramsAndTypes := make([]string, 0, len(params))

	for key := range params {
		paramsAndTypes = append(paramsAndTypes, key+" "+params[key])
	}

	paramsText := "id INTEGER PRIMARY KEY, " + strings.Join(paramsAndTypes[:], ", ")
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( %s )", tableName, paramsText)

	statement, _ := db.Prepare(query)
	statement.Exec()

}

func InsertToTable(db *sqlx.DB, tableName string, data map[string]interface{}) {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for key := range data {
		columns = append(columns, key)
		values = append(values, data[key])
	}

	columnsText := strings.Join(columns[:], ", ")
	valuesText := "?" + strings.Repeat(", ?", len(data)-1)

	query := "INSERT INTO " + tableName + " ( " + columnsText + " ) VALUES ( " + valuesText + " )"
	stmt, err := db.Prepare(query)

	if err != nil {
		fmt.Println(err)
	}

	stmt.Exec(values...)
}

func QueryDatabase(db *sqlx.DB, tableName string) *sqlx.Rows {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Queryx(query)

	if err != nil {
		panic(err)
	}
	return rows
}
