package main

import "time"

type User struct {
	Id        int64     `json:"id"`
	Photo     string    `json:"photo"`
	Name      string    `json:"name"`
	Alias     string    `json:"alias"`
	Password  string    `json:"password"`
	DeletedAt time.Time `json:"deletedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MARK: - Storage

func CreateUserTable(db DbClient) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			photo TEXT,
			name TEXT NOT NULL,
			alias TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			deleted_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		)

		CREATE TRIGGER update_user_updated_at 
		AFTER UPDATE ON users
		BEGIN
			UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
		`)
	if err != nil {
		return err
	}
	return nil
}

// SaveUser creates a new user in the database.
func SaveUser(db DbClient, fullName, alias, plainPassword string) error {
	_, err := db.Exec(`
		INSERT INTO users (name, alias, password) VALUES (?, ?, ?)
	`, fullName, alias, HashPassword(plainPassword))
	if err != nil {
		return err
	}
	return nil
}

func GetUser(db DbClient, id int) (User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, photo, name, alias, password, created_at, updated_at 
		FROM users
		WHERE id = ?
	`, id).Scan(&user.Id, &user.Photo, &user.Name, &user.Alias, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByAlias(db DbClient, alias string) (User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, photo, name, alias, password, created_at, updated_at 
		FROM users
		WHERE alias = ?
	`, alias).Scan(&user.Id, &user.Photo, &user.Name, &user.Alias, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

// SoftDeleteUser marks a user as deleted.
func SoftDeleteUser(db DbClient, id int) error {
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
func RemoveUser(db DbClient, id int) error {
	_, err := db.Exec(`
		DELETE FROM users WHERE id = ?
	`, id)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUserPhoto(db DbClient, id int, photo string) error {
	_, err := db.Exec(`
		UPDATE users SET photo = ? WHERE id = ?
	`, photo, id)
	if err != nil {
		return err
	}
	return nil
}