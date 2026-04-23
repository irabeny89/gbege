package main

import (
	"database/sql"
	"errors"
	"net/url"
	"os"

	_ "modernc.org/sqlite"
)

// MARK: - Structs

// DbClient is a client that is used for reading and writing to the database.
type DbClient struct {
	// Read pool (multi connections). E.g. SELECT.
	readPool *sql.DB
	// Write pool (single connection). E.g. INSERT, UPDATE, DELETE.
	writePool *sql.DB
}

type sqliteConfig struct {
	// dbPath is the path to the database file E.g "app.db".
	dbPath string
	// driver is the driver to use for the database connection. E.g "sqlite".
	driver string
	// url is the Data Source Name used to connect to the database. E.g ":memory:" or "sqlite:///app.db"
	url string
	// pragmas are the pragmas to use for the database connection. E.g []string{"journal_mode(WAL)", "busy_timeout(5000)", "foreign_keys(ON)"}
	pragma []string
	// mode is the mode to use for the database connection (e.g. "ro", "rw", "rwc" & "memory". Default is "rwc" ).
	mode string
}

// MARK: - Interfaces

// MARK: - Const & Var

// MARK: - Methods

// newPool creates a new database connection pool for sqlite (or other drivers)
func (cnf *sqliteConfig) newPool(maxConn int) (*sql.DB, error) {
	if cnf.driver == "" {
		return nil, errors.New("driver is not set")
	}
	if cnf.url != "" {
		return sql.Open(cnf.driver, cnf.url)
	}
	var (
		scheme = "file"
	)
	if cnf.dbPath == "" {
		cnf.dbPath = "app.db"
	}
	if cnf.mode == "" {
		cnf.mode = "rwc"
	}
	query := url.Values{
		"mode": []string{cnf.mode},
	}
	if len(cnf.pragma) > 0 {
		query["_pragma"] = cnf.pragma
	}
	url := url.URL{
		Scheme:   scheme,
		Path:     cnf.dbPath,
		RawQuery: query.Encode(),
	}
	db, err := sql.Open(cnf.driver, url.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConn)
	return db, nil
}

// Ping checks if the database connection is alive.
func (c *DbClient) Ping() error {
	err := c.readPool.Ping()
	if err != nil {
		return err
	}
	err = c.writePool.Ping()
	if err != nil {
		return err
	}
	return nil
}
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

// MARK: - Private Func

// MARK: - Public Func

// NewDbClient creates a new database client.
func NewDbClient() (*DbClient, error) {
	var (
		dbPath = "app.db"
		driver = "sqlite"
		pragma = []string{
			"journal_mode(WAL)",
			"busy_timeout(5000)",
			"foreign_keys(ON)",
		}
	)
	if dbPathEnv, ok := os.LookupEnv("DB_PATH"); ok {
		dbPath = dbPathEnv
	}
	readConfig := &sqliteConfig{
		dbPath: dbPath,
		driver: driver,
		mode:   "ro",
		pragma: pragma,
	}
	writeConfig := &sqliteConfig{
		dbPath: dbPath,
		driver: driver,
		mode:   "rwc",
		pragma: append(pragma, "synchronous(NORMAL)"),
	}

	//! NOTE: sqlite WAL mode requires -shm and -wal files.
	//! Initialize the writer connection first (1 connection).
	//! This is because WAL mode requires the "writer" connection (rwc) to create the -shm and -wal files,
	//! and the "reader" connection (ro) cannot create them.

	w, err := writeConfig.newPool(1)
	if err != nil {
		return nil, err
	}
	r, err := readConfig.newPool(10)
	if err != nil {
		return nil, err
	}

	client := &DbClient{
		readPool:  r,
		writePool: w,
	}

	return client, nil
}
