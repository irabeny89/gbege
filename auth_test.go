package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewToken(t *testing.T) {
	token1 := NewToken()
	token2 := NewToken()

	if len(token1) == 0 || len(token2) == 0 {
		t.Error("NewToken returned empty string")
	}

	if bytes.Equal(token1, token2) {
		t.Error("NewToken generated duplicate tokens")
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "my-super-secret-password"

	hash := HashPassword(password)
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Hash does not have the expected prefix. Got: %s", hash)
	}

	// Verify with correct password
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned an error for a valid password: %v", err)
	}
	if !match {
		t.Error("VerifyPassword returned false for the correct password")
	}

	// Verify with wrong password
	match, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned an error for an invalid password: %v", err)
	}
	if match {
		t.Error("VerifyPassword returned true for an incorrect password")
	}
}

func TestDecodeHash(t *testing.T) {
	// Invalid format (completely disjoint)
	invalidHash := "invalid-hash-string"
	_, err := DecodeHash(invalidHash)
	if err == nil {
		t.Error("DecodeHash should have returned an error for invalid format")
	}

	// Invalid format (missing expected parts)
	invalidHash2 := "$argon2id$v=19$m=65536,t=3,p=2$salt"
	_, err = DecodeHash(invalidHash2)
	if err == nil {
		t.Error("DecodeHash should have returned an error for incorrect part count")
	}

	// Valid hash generated using HashPassword
	hash := HashPassword("test-password")

	params, err := DecodeHash(hash)
	if err != nil {
		t.Errorf("DecodeHash returned an error for a valid hash: %v", err)
	}

	if params.Memory == 0 || params.Iterations == 0 || params.Parallelism == 0 {
		t.Error("DecodeHash failed to correctly parse parameters")
	}

	if len(params.Salt) == 0 || len(params.Hash) == 0 {
		t.Error("DecodeHash failed to correctly parse salt and hash")
	}
}
