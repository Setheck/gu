package gu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name    string
		slice   []string
		str     string
		inslice bool
	}{
		{"only_one", []string{"one"}, "one", true},
		{"first", []string{"one", "two"}, "one", true},
		{"last", []string{"a", "b", "c", "one"}, "one", true},
		{"empty slice", []string{}, "one", false},
		{"empty str", []string{"test"}, "", false},
		{"both empty", []string{}, "", false},
		{"nil slice", nil, "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringInSlice(test.slice, test.str)
			assert.Equal(t, test.inslice, result)
		})
	}
}

func TestInt32InSlice(t *testing.T) {
	tests := []struct {
		name    string
		slice   []int32
		integer int32
		inslice bool
	}{
		{"only_one", []int32{1}, 1, true},
		{"first", []int32{1, 2}, 1, true},
		{"last", []int32{0, 4, 7, 10}, 10, true},
		{"empty slice", []int32{}, 1, false},
		{"nil slice", nil, 7, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Int32InSlice(test.slice, test.integer)
			assert.Equal(t, test.inslice, result)
		})
	}
}

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
		lastValue = str
	}
}

func TestBase64Encoding(t *testing.T) {
	str := "test value... atsu atsu atsu 🔥🔥🔥🔥🔥🔥🔥🔥🔥"
	encoded := EncodeBase64(str)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, str, string(decoded))
}
