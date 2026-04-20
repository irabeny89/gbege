package main

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

func TestUserLifecycle(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	client := DbClient{readPool: db, writePool: db}

	err := CreateUserTable(client)
	if err != nil {
		t.Fatalf("CreateUserTable failed: %v", err)
	}

	fullName := "John Doe"
	alias := "johndoe"
	plainPassword := "secretpassword"

	err = SaveUser(client, fullName, alias, plainPassword)
	if err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	userByAlias, err := GetUserByAlias(client, alias)
	if err != nil {
		t.Fatalf("GetUserByAlias failed: %v", err)
	}

	if userByAlias.Name != fullName {
		t.Errorf("Expected name %s, got %s", fullName, userByAlias.Name)
	}

	userById, err := GetUser(client, int(userByAlias.Id))
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if userById.Id != userByAlias.Id {
		t.Errorf("Expected id %d, got %d", userByAlias.Id, userById.Id)
	}

	// Test UpdateUserPhoto
	newPhoto := "https://example.com/photo.jpg"
	err = UpdateUserPhoto(client, int(userById.Id), newPhoto)
	if err != nil {
		t.Fatalf("UpdateUserPhoto failed: %v", err)
	}

	updatedUser, _ := GetUser(client, int(userById.Id))
	if updatedUser.Photo != newPhoto {
		t.Errorf("Expected photo %s, got %s", newPhoto, updatedUser.Photo)
	}

	// Test SoftDeleteUser
	err = SoftDeleteUser(client, int(userById.Id))
	if err != nil {
		t.Fatalf("SoftDeleteUser failed: %v", err)
	}

	deletedUser, _ := GetUser(client, int(userById.Id))
	if deletedUser.DeletedAt.IsZero() {
		t.Error("Expected DeletedAt to be set")
	}

	// Test RemoveUser
	err = RemoveUser(client, int(userById.Id))
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}

	_, err = GetUser(client, int(userById.Id))
	if err == nil {
		t.Error("Expected error when getting removed user, but got nil")
	}
}
