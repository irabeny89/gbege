package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	Salt        []byte
	Hash        []byte
}

const (
	memory      = 64 * 1024 // 64MB
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

// NewToken generates a new random token.
func NewToken() []byte {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

// DecodeHash decodes a hash string into a Params struct.
func DecodeHash(encodedHash string) (*Params, error) {
	//? Standard PHC format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, errors.New("invalid hash format")
	}

	var p Params
	// 1. Version check (optional)
	// vals[1] is "argon2id", vals[2] is "v=19"

	// 2. Parse Parameters (m, t, p)
	_, err := fmt.Sscanf(
		vals[3],
		"m=%d,t=%d,p=%d",
		&p.Memory,
		&p.Iterations,
		&p.Parallelism,
	)
	if err != nil {
		return nil, err
	}

	// 3. Decode Salt
	p.Salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, err
	}

	// 4. Decode Hash
	p.Hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// HashPassword hashes a password.
func HashPassword(password string) string {
	// 1. Generate a cryptographically secure random salt
	salt := make([]byte, saltLength)
	rand.Read(salt)

	// 2. Generate the hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// 3. Encode to PHC format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	// We use RawStdEncoding to avoid the '=' padding characters
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash
}

// VerifyPassword verifies a password against a salt and hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	p, err := DecodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Re-hash the login attempt with the extracted salt and params
	comparisonHash := argon2.IDKey(
		[]byte(password),
		p.Salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		uint32(len(p.Hash)),
	)

	// Use subtle.ConstantTimeCompare to avoid timing attacks
	if subtle.ConstantTimeCompare(p.Hash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}
