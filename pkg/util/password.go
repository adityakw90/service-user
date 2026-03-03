package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateRandomPassword generates a random password with the given length.
// The password will contain at least one uppercase letter, one lowercase letter, one digit, and one special character.
func GenerateRandomPassword(length int) (string, error) {
	if length < 4 {
		return "", fmt.Errorf("password length must be at least 4")
	}

	upperCase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerCase := "abcdefghijklmnopqrstuvwxyz"
	digits := "0123456789"
	specialChars := "!@#$%^&*()-_=+[]{}|;:,.<>?/"

	allChars := upperCase + lowerCase + digits + specialChars

	// Ensure the password contains at least one character from each set
	password := make([]byte, length)
	charSets := []string{upperCase, lowerCase, digits, specialChars}

	// Select one character from each required set
	for i, set := range charSets {
		char, err := randomChar(set)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Fill the remaining characters randomly
	for i := 4; i < length; i++ {
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	// Shuffle the password to avoid predictable order
	shuffle(password)

	return string(password), nil
}

// randomChar selects a random character from the given character set.
func randomChar(charset string) (byte, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[index.Int64()], nil
}

// shuffle randomizes the order of characters in a byte slice.
func shuffle(password []byte) {
	for i := len(password) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
}
