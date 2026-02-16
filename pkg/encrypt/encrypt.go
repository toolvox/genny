// Package encrypt provides AES-256-GCM encryption with PBKDF2-SHA256 key derivation
// for client-side decryptable page encryption.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

const (
	saltSize   = 16
	ivSize     = 12
	keySize    = 32 // AES-256
	iterations = 100_000
)

// Encrypt encrypts plaintext using AES-256-GCM with a key derived from passphrase via PBKDF2-SHA256.
// Returns the salt, IV, and ciphertext (with GCM auth tag appended).
func Encrypt(plaintext, passphrase string) (salt, iv, ciphertext []byte, err error) {
	salt = make([]byte, saltSize)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	iv = make([]byte, ivSize)
	if _, err = rand.Read(iv); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	key := pbkdf2([]byte(passphrase), salt, iterations, keySize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertext = gcm.Seal(nil, iv, []byte(plaintext), nil)
	return salt, iv, ciphertext, nil
}

// pbkdf2 derives a key using PBKDF2-HMAC-SHA256.
func pbkdf2(password, salt []byte, iter, keyLen int) []byte {
	numBlocks := (keyLen + sha256.Size - 1) / sha256.Size
	var dk []byte

	for block := 1; block <= numBlocks; block++ {
		dk = append(dk, pbkdf2Block(password, salt, iter, block)...)
	}
	return dk[:keyLen]
}

func pbkdf2Block(password, salt []byte, iter, blockNum int) []byte {
	h := hmac.New(sha256.New, password)

	// U1 = PRF(password, salt || INT_32_BE(blockNum))
	h.Write(salt)
	h.Write([]byte{byte(blockNum >> 24), byte(blockNum >> 16), byte(blockNum >> 8), byte(blockNum)})
	u := h.Sum(nil)
	result := make([]byte, len(u))
	copy(result, u)

	// U2..Uiter
	for i := 1; i < iter; i++ {
		h.Reset()
		h.Write(u)
		u = h.Sum(nil)
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}
