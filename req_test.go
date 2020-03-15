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

func TestSamplePercentile(t *testing.T) {
	var values []int64
	for i := 0; i < 100; i++ {
		for j := 1; j <= 10; j++ {
			values = append(values, int64(j))
		}
	}

	values = []int64{}
	for i := 0; i < 1000; i++ {
		values = append(values, int64(i))
	}

	want := []float64{
		0, 99.10000000000001, 499.5, 599.6, 699.6999999999999, 799.8000000000001,
		899.9, 949.9499999999999, 999,
	}
	got := make([]float64, 0)
	for _, perc := range []float64{0, 0.1, 0.5, .6, .7, .8, .9, .95, 1} {
		got = append(got, SamplePercentile(int64Slice(values), perc))
	}
	assert.Equal(t, want, got)
}
