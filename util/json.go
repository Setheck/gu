package util

import (
	"bytes"
	"encoding/json"
)

// JsonPrettyPrint attempts to reformat a json byte array to be pretty.
// if any error occurs, just fall back to the original input.
func JsonPrettyPrint(in []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "\t")
	if err != nil {
		return in
	}
	return out.Bytes()
}
