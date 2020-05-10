package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		idx   int
	}{
		{"only_one", []string{"one"}, "one", 0},
		{"first", []string{"one", "two"}, "one", 0},
		{"last", []string{"a", "b", "c", "one"}, "one", 3},
		{"empty slice", []string{}, "one", -1},
		{"empty str", []string{"test"}, "", -1},
		{"both empty", []string{}, "", -1},
		{"nil slice", nil, "", -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := StringInSlice(test.slice, test.str)
			assert.Equal(t, test.idx, result)
		})
	}
}

func TestInt32InSlice(t *testing.T) {
	tests := []struct {
		name    string
		slice   []int32
		integer int32
		idx     int
	}{
		{"only_one", []int32{1}, 1, 0},
		{"first", []int32{1, 2}, 1, 0},
		{"last", []int32{0, 4, 7, 10}, 10, 3},
		{"empty slice", []int32{}, 1, -1},
		{"nil slice", nil, 7, -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Int32InSlice(test.slice, test.integer)
			assert.Equal(t, test.idx, result)
		})
	}
}
