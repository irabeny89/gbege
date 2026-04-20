package main

import (
	"database/sql"
	"net/url"

	_ "modernc.org/sqlite"
)

// MARK: - Type, Const & Var

// DbClient is a client that is used for reading and writing to the database.
type DbClient struct {
	// Read pool (multi connections). E.g. SELECT.
	readPool *sql.DB
	// Write pool (single connection). E.g. INSERT, UPDATE, DELETE.
	writePool *sql.DB
}

var dbPath = "app.db"

const driver = "sqlite"
const scheme = "file"

var commonPragma = []string{
	"journal_mode(WAL)",
	"busy_timeout(5000)",
	"foreign_keys(ON)",
}

// MARK: - Methods

// Query executes a query that returns rows, using the read pool. E.g SELECT * FROM users
func (c *DbClient) Query(query string, args ...any) (*sql.Rows, error) {
	return c.readPool.Query(query, args...)
}

// QueryRow executes a query that returns a single row, using the read pool. E.g SELECT * FROM users WHERE id = 1
func (c *DbClient) QueryRow(query string, args ...any) *sql.Row {
	return c.readPool.QueryRow(query, args...)
}

// Exec executes a query that returns a result, using the write pool. E.g INSERT, UPDATE, DELETE, CREATE, DROP, etc
func (c *DbClient) Exec(query string, args ...any) (sql.Result, error) {
	return c.writePool.Exec(query, args...)
}

// Close gracefully closes all db pools
func (c *DbClient) Close() error {
	c.readPool.Close()
	c.writePool.Close()
	return nil
}

// MARK: - Private Func

// newReadPool creates a read-only database handle.
func newReadPool() (*sql.DB, error) {
	query := url.Values{
		"_pragma": commonPragma,
		// read-only mode to prevent accidental writes for db reader.
		"mode": []string{"ro"},
	}
	url := url.URL{
		Scheme:   scheme,
		Path:     dbPath,
		RawQuery: query.Encode(),
	}
	db, err := sql.Open(driver, url.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(50)

	return db, nil
}

// newWritePool creates a database handle for writing to the database.
func newWritePool() (*sql.DB, error) {
	query := url.Values{
		// normal sync mode for db writer.
		"_pragma": append(commonPragma, "synchronous(NORMAL)"),
	}
	url := url.URL{
		Scheme:   scheme,
		Path:     dbPath,
		RawQuery: query.Encode(),
	}
	db, err := sql.Open(driver, url.String())
	if err != nil {
		return nil, err
	}
	// Limits to 1 connection pool to prevent "database is locked" errors.
	db.SetMaxOpenConns(1)

	return db, nil
}

func NewDbClient() (*DbClient, error) {
	//! NOTE: sqlite WAL mode requires -shm and -wal files.
	//! Initialize the writer connection first
	//! this creates -shm and -wal files because the reader(ro) cannot create them.

	w, wErr := newWritePool()
	if wErr != nil {
		return nil, wErr
	}
	r, rErr := newReadPool()
	if rErr != nil {
		return nil, rErr
	}

	client := &DbClient{
		readPool:  r,
		writePool: w,
	}

	return client, nil
}

