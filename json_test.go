package gu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonPrettyPrint(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"pretty",
			`{"one":"two","map":{"key":"val"},"ary":["a","b"]}`,
			"{\n\t\"one\": \"two\",\n\t\"map\": {\n\t\t\"key\": \"val\"\n\t},\n\t\"ary\": [\n\t\t\"a\",\n\t\t\"b\"\n\t]\n}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := JsonPrettyPrint([]byte(test.input))
			assert.Equal(t, test.want, string(got))

			// For visual inspection
			fmt.Printf("input:\n%v\n", test.input)
			fmt.Printf("output:\n%v\n", string(got))
		})
	}
}
