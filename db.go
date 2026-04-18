package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

// NewDB creates a new database connection.
func NewDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./gbege.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}

