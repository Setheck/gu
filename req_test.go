package gu

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestMarshalWithPretty(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		pretty string
		want   string
	}{
		{"not pretty", struct {
			One string `json:"one"`
		}{
			"two",
		}, "", `{"one":"two"}`},
		{"not pretty-false", struct {
			One string `json:"one"`
		}{
			"two",
		}, "false", `{"one":"two"}`},
		{"not pretty-0", struct {
			One string `json:"one"`
		}{
			"two",
		}, "0", `{"one":"two"}`},
		{"pretty-true", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
		{"pretty-true", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
		{"pretty-1", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("pretty=%s", test.pretty),
				},
			}
			got := MarshalWithPretty(r, test.input)
			assert.Equal(t, test.want, string(got))
		})
	}
}
