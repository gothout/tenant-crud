package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength  = 16        // 128-bit salt
	memory      = 64 * 1024 // 64 MB memory usage
	iterations  = 3         // Number of iterations
	parallelism = 2         // Number of threads
	keyLength   = 32        // 256-bit derived key
)

// Hash generates a secure Argon2id hash from the given password.
// The returned string includes algorithm parameters, base64-encoded salt, and hash.
func Hash(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	return encoded, nil
}

// Compare checks whether the given password matches the provided Argon2id hash.
// It returns nil if the password is correct or an error otherwise.
func Compare(encodedHash, password string) error {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return errors.New("invalid hash format")
	}

	var mem uint32
	var iter uint32
	var par uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iter, &par)
	if err != nil {
		return fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, iter, mem, par, uint32(len(expectedHash)))

	if !constantTimeCompare(expectedHash, computedHash) {
		return errors.New("invalid password")
	}

	return nil
}

// constantTimeCompare performs a constant-time comparison between two byte slices
// to prevent timing attacks.
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
