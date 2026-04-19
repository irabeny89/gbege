package main

import (
	"database/sql"
	"net/url"

	_ "modernc.org/sqlite"
)

// DbWriter is a wrapper around sql.DB that is used for writing to the database.
type DbWriter struct {
	*sql.DB
}

// DbReader is a wrapper around sql.DB that is used for reading from the database.
type DbReader struct {
	*sql.DB
}

// DbPool is a pool of database connections.
type DbPool struct {
	// Writer pool (single connection). E.g. INSERT, UPDATE, DELETE.
	Writer *DbWriter
	// Reader pool (multi connections). E.g. SELECT.
	Reader *DbReader
}

type DbConn struct {
	driver string
	url    string
}

const dbPath = "app.db"
const driver = "sqlite"

var commonPragma = []string{
	"journal_mode(WAL)",
	"busy_timeout(5000)",
	"foreign_keys(ON)",
}

// newDbReader creates a read-only database handle.
func newDbReader() (*DbReader, error) {
	// read-only mode to prevent accidental writes for db reader.
	readerMode := []string{"ro"}
	readerQuery := url.Values{
		"_pragma": commonPragma,
		"mode":    readerMode,
	}
	readerUrl := url.URL{
		Scheme:   "file",
		Path:     dbPath,
		RawQuery: readerQuery.Encode(),
	}
	reader, err := sql.Open(driver, readerUrl.String())
	if err != nil {
		return nil, err
	}
	reader.SetMaxOpenConns(0) // unlimited reads

	return &DbReader{reader}, nil
}

// newDbWriter creates a database handle for writing to the database.
func newDbWriter() (*DbWriter, error) {
	// normal sync mode for db writer.
	writerPragma := []string{"synchronous(NORMAL)"}
	writerQuery := url.Values{
		"_pragma": append(writerPragma, commonPragma...),
	}
	writerUrl := url.URL{
		Scheme:   "file",
		Path:     dbPath,
		RawQuery: writerQuery.Encode(),
	}
	writer, wErr := sql.Open(driver, writerUrl.String())
	if wErr != nil {
		return nil, wErr
	}
	// Limits to 1 connection pool to prevent "database is locked" errors.
	writer.SetMaxOpenConns(1)

	return &DbWriter{writer}, nil
}

// NewDB creates a new database connection with default config.
func NewDB() (*sql.DB, error) {
	db, err := sql.Open(driver, dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewDbPool creates a new database connection pool with custom configs.
func NewDbPool() (*DbPool, error) {
	//! NOTE: sqlite WAL mode requires -shm and -wal files.
	//! Initialize the writer connection first
	//! this creates -shm and -wal files because the reader(ro) cannot create them.

	writer, wErr := newDbWriter()
	if wErr != nil {
		return nil, wErr
	}
	reader, rErr := newDbReader()
	if rErr != nil {
		return nil, rErr
	}

	return &DbPool{
		Reader: reader,
		Writer: writer,
	}, nil
}
