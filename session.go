package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Session struct {
	Id        []byte    `json:"id"`
	UserId    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// time to live for a session is 30 days
const ttl = 30 * 24 * time.Hour

// MARK: - Storage

// CreateSessionTable creates the sessions table in the database.
func CreateSessionTable(db *sql.DB) error {
	_, err := db.Exec(
		`
		CREATE TABLE IF NOT EXISTS sessions (
			id BLOB PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
		`,
	)
	if err != nil {
		return err
	}

	return nil
}

// SaveSession creates a new session for a user in the database.
func SaveSession(db *sql.DB, userId int) error {
	t := time.Now()
	_, err := db.Exec(
		`
		INSERT INTO sessions (id, user_id, expires_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
		`,
		NewToken(), userId, t.Add(ttl), t, t,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetSession retrieves a session from the database by its ID.
func GetSession(db *sql.DB, id []byte) (Session, error) {
	var session Session
	err := db.QueryRow(
		`
		SELECT id, user_id, expires_at, created_at, updated_at 
		FROM sessions
		WHERE id = ?
		`,
		id,
	).Scan(&session.Id, &session.UserId, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return session, err
	}

	return session, nil
}

// TODO: create a job to run this everyday
// DeleteSession deletes a session from the database.
func DeleteSession(db *sql.DB, id []byte) error {
	_, err := db.Exec(
		`
		DELETE FROM sessions
		WHERE id = ?
		`,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteExpiredSessions deletes all expired sessions from the database.
func DeleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec(
		`
		DELETE FROM sessions
		WHERE expires_at < ?
		`,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}
