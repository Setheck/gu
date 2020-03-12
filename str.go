package gu

import (
	"encoding/base64"
	"math/rand"
	"strings"
	"time"
)

func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

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

func Int32InSlice(slice []int32, val int32) bool {
	if slice == nil {
		return false
	}
	for _, item := range slice {
		if val == item {
			return true
		}
	}
	return false
}

func StringInSlice(slice []string, str string) bool {
	if slice == nil {
		return false
	}
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
