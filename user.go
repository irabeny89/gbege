package main

import "time"

type User struct {
	Id        int64     `json:"id"`
	FullName  string    `json:"fullName"`
	Alias     string    `json:"alias"`
	Password  string    `json:"password"`
	DeletedAt time.Time `json:"deletedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MARK: - Storage

func CreateUserTable(db DbWriter) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			alias TEXT NOT NULL,
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

func SaveUser(db DbWriter, fullName, alias, plainPassword string) error {
	_, err := db.Exec(`
		INSERT INTO users (full_name, alias, password) VALUES (?, ?, ?)
	`, fullName, alias, HashPassword(plainPassword))
	if err != nil {
		return err
	}
	return nil
}

func GetUser(db DbReader, id int) (User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, full_name, alias, password, created_at, updated_at 
		FROM users
		WHERE id = ?
	`, id).Scan(&user.Id, &user.FullName, &user.Alias, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
