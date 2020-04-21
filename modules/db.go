package modules

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite3 driver
)

var dbConn *sql.DB

func loadSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./storage.db?cache=shared&mode=memory")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func loadMySQL(settings map[string]interface{}) (*sql.DB, error) {
	host := settings["host"].(string)
	port := settings["port"].(string)
	user := settings["user"].(string)
	password := settings["password"].(string)
	schema := settings["schema"].(string)

	db, err := sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/"+schema)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// GetDB returns a DB connection instance
func GetDB(settings map[string]interface{}) (*sql.DB, error) {
	var err error

	if len(settings) == 0 {
		dbConn, err = loadSQLite()
	} else {
		dbConn, err = loadMySQL(settings)
	}

	return dbConn, err
}

// CloseDB closes the DB connection
func CloseDB() {
	dbConn.Close()
}
