package util

import (
	"encoding/base64"
	"math/rand"
	"strings"
	"time"
)

// EncodeBase64 is a simple string to string wrapper for encoding a string to the base64 counterpart
func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// DecodeBase64 is a simple string to string wrapper for decoding a string from base64
func DecodeBase64(input string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(input)
	return string(b), err
}

// RandomString simply generate a random string of letters with given length
func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]string, 0, length)
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(charSet))
		result = append(result, string(charSet[idx]))
	}
	return strings.Join(result, "")
}
