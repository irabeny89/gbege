package user

import (
	"database/sql"
	"time"

	"github.com/irabeny89/gbege/internal/security"
	"github.com/irabeny89/gosqlitex"
)

// MARK: - Type, Const & Var

type User struct {
	Id        int64          `json:"id"`
	Photo     sql.NullString `json:"photo,omitempty"`
	Username  string         `json:"name"`
	Password  string         `json:"-"`
	DeletedAt sql.NullTime   `json:"deletedAt,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// MARK: - Save
// SaveUser creates a new user in the database.
func SaveUser(db *gosqlitex.DbClient, username, plainPassword string) (*User, error) {
	res, err := db.Exec(`
		INSERT INTO users (username, password)
		VALUES (?, ?)
	`, username, security.HashPassword(plainPassword))

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetUser(db, int(id))
}

// MARK: - Get
// GetUser retrieves a user from the database by their ID.
func GetUser(db *gosqlitex.DbClient, id int) (*User, error) {
	u := new(User)
	err := db.QueryRow(`
		SELECT id, COALESCE(photo, ''), username, password, deleted_at, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&u.Id, &u.Photo, &u.Username, &u.Password, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// If the user is deleted(soft), return nil
	if u.DeletedAt.Valid {
		return nil, nil
	}
	return u, nil
}

// MARK: - By username
// GetUserByUsername retrieves a user from the database by their alias.
func GetUserByUsername(db *gosqlitex.DbClient, alias string) (*User, error) {
	u := new(User)
	err := db.QueryRow(`
		SELECT id, COALESCE(photo, ''), username, password, deleted_at, created_at, updated_at
		FROM users
		WHERE username = ?
	`, alias).Scan(&u.Id, &u.Photo, &u.Username, &u.Password, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// If the user is deleted(soft), return nil
	if u.DeletedAt.Valid {
		return nil, nil
	}
	return u, nil
}

// SoftDeleteUser marks a user as deleted.
func SoftDeleteUser(db *gosqlitex.DbClient, id int) error {
	_, err := db.Exec(`
		UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
	`, id)
	if err != nil {
		return err
	}
	return nil
}

// RemoveUser deletes a user from the database permanently.
// This should be used with caution as it will delete all associated data.
func RemoveUser(db *gosqlitex.DbClient, id int) error {
	_, err := db.Exec(`
		DELETE FROM users WHERE id = ?
	`, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserPhoto updates the photo of a user.
func UpdateUserPhoto(db *gosqlitex.DbClient, id int, photo string) error {
	_, err := db.Exec(`
		UPDATE users SET photo = ? WHERE id = ?
	`, photo, id)
	if err != nil {
		return err
	}
	return nil
}

// CleanupDeletedUsers removes users that have been soft deleted for more than 6 months.
func CleanupDeletedUsers(db *gosqlitex.DbClient) error {
	_, err := db.Exec(`
			DELETE FROM users where deleted_at < DATE('now', '-6 months');
		`)
	if err != nil {
		return err
	}
	return nil
}
