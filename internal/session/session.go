package session

import (
	"database/sql"
	"time"

	"github.com/irabeny89/gbege/internal/security"
	"github.com/irabeny89/gosqlitex"
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

// MARK: - Save
// SaveSession creates a new session for a user in the database.
func SaveSession(db *gosqlitex.DbClient, userId int) (*Session, error) {
	t := time.Now()
	id := security.NewToken()
	_, err := db.Exec(
		`
		INSERT INTO sessions (id, user_id, expires_at) 
		VALUES (?, ?, ?)
		`,
		id, userId, t.Add(ttl),
	)

	if err != nil {
		return nil, err
	}

	return GetSession(db, id)
}

// MARK: - Get
// GetSession retrieves a session from the database by its ID.
func GetSession(db *gosqlitex.DbClient, id []byte) (*Session, error) {
	s := new(Session)
	err := db.QueryRow(
		`
		SELECT id, user_id, expires_at, created_at, updated_at 
		FROM sessions
		WHERE id = ?
		`,
		id,
	).Scan(&s.Id, &s.UserId, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return s, nil
}

// MARK: - By user id
// GetSessionByUserId retrieves a session from the database by user ID
func GetSessionByUserId(db *gosqlitex.DbClient, userId int) (*Session, error) {
	s := new(Session)
	err := db.QueryRow(
		`
		SELECT id, user_id, expires_at, created_at, updated_at 
		FROM sessions
		WHERE user_id = ?
		`,
		userId,
	).Scan(&s.Id, &s.UserId, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return s, nil
}

// MARK: - Delete
// DeleteSession deletes a session from the database.
func DeleteSession(db *gosqlitex.DbClient, id []byte) error {
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

// MARK: - DeleteExpiredSessions
// DeleteExpiredSessions deletes all expired sessions from the database.
func DeleteExpiredSessions(db *gosqlitex.DbClient) error {
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
