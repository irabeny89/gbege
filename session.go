package main

import (
	"time"

	_ "modernc.org/sqlite"
)

type Session struct {
	Id        []byte    `json:"id"`
	UserId    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// time to live for a session is 30 days
const ttl = 30 * 24 * time.Hour

// MARK: - Storage

// CreateSessionTable creates the sessions table in the database.
func CreateSessionTable(db DbClient) error {
	_, err := db.Exec(
		`
		CREATE TABLE IF NOT EXISTS sessions (
			id BLOB PRIMARY KEY,
			user_id INTEGER NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
			ON DELETE CASCADE
			ON UPDATE CASCADE
		)

		CREATE TRIGGER update_session_updated_at 
		AFTER UPDATE ON sessions
		BEGIN
			UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
		`,
	)
	if err != nil {
		return err
	}

	return nil
}

// SaveSession creates a new session for a user in the database.
func SaveSession(db DbClient, userId int) error {
	t := time.Now()
	_, err := db.Exec(
		`
		INSERT INTO sessions (id, user_id, expires_at) 
		VALUES (?, ?, ?)
		`,
		NewToken(), userId, t.Add(ttl),
	)
	if err != nil {
		return err
	}

	return nil
}

// GetSession retrieves a session from the database by its ID.
func GetSession(db DbClient, id []byte) (Session, error) {
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
func DeleteSession(db DbClient, id []byte) error {
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
func DeleteExpiredSessions(db DbClient) error {
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
