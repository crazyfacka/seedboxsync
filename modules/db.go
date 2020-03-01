package modules

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

var dbConn *sql.DB

// GetDB returns a DB connection instance
func GetDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:storage.db?cache=shared&mode=memory")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	dbConn = db
	return db, nil
}

// CloseDB closes the DB connection
func CloseDB() {
	dbConn.Close()
}
