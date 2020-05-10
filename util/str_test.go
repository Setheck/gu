package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for i := 0; i < 100; i++ {
		lastValue := ""
		str := RandomString(i)
		assert.Len(t, str, i)
		if i > 0 {
			assert.NotEmpty(t, str)
			assert.NotEqual(t, lastValue, str)
		} else {
			assert.Empty(t, str)
		}
	}
}

func TestBase64Encoding(t *testing.T) {
	str := "test value... atsu atsu atsu 🔥🔥🔥🔥🔥🔥🔥🔥🔥"
	encoded := EncodeBase64(str)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, str, decoded)
}
