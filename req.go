package gu

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// MarshalWithPretty for use in handlers, check for the 'pretty' query parameter
// if the pretty parameter is true, pretty json will be returned.
func MarshalWithPretty(r *http.Request, obj interface{}) []byte {
	p := r.URL.Query().Get("pretty")
	out, _ := json.Marshal(obj)
	if pretty, _ := strconv.ParseBool(p); pretty {
		out = JsonPrettyPrint(out)
	}
	return out
}
