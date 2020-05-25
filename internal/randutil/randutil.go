package randutil

// based on https://www.calhoun.io/creating-random-strings-in-go/

import (
	"math/rand"
	"time"
)

const (
	// Uppercase contains all uppercase ascii characters
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Lowercase contains all lowercase ascii characters
	Lowercase = "abcdefghijklmnopqrstuvwxyz"

	// Digits contains all digits from 0 to 9
	Digits = "01234567890"
)

const charset = Uppercase + Lowercase + Digits

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
