/*
	Package pki implements basic functions for managing the pki
	represented as a SQL database.
*/

package pki

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
)

func OpenDatabase(dataSourceName, dbDriver string) (*sqlx.DB, error) {

	var db *sqlx.DB
	db, err := sqlx.Connect(dbDriver, dataSourceName)

	if err != nil {
		return nil, err
	}

	return db, err
}

func CreateTable(db *sqlx.DB, tableName string, params map[string]string) error{
	paramsAndTypes := make([]string, 0, len(params))

	for key := range params {
		paramsAndTypes = append(paramsAndTypes, key+" "+params[key])
	}

	paramsText := "idx INTEGER PRIMARY KEY, " + strings.Join(paramsAndTypes[:], ", ")
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ( %s )", tableName, paramsText)

	statement, err := db.Prepare(query)

	if err != nil{
		return err
	}
	_, err = statement.Exec()
	if err != nil{
		return err
	}

	return nil

}


func InsertIntoTable(db *sqlx.DB, tableName string, id, typ string, config []byte) error{
	query :="INSERT INTO " + tableName + " (Id, Typ, Config) VALUES (?, ?, ?)"

	stmt, err := db.Prepare(query)
	if err != nil{
		return err
	}
	_, err = stmt.Exec(id, typ, config)
	if err != nil{
		return err
	}

	return nil
}

func QueryDatabase(db *sqlx.DB, tableName string, condition string) (*sqlx.Rows, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE Typ = ?", tableName)
	rows, err := db.Queryx(query, condition)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func rowExists(db *sqlx.DB, query string, args ...interface{}) (bool, error) {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := db.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}