package util

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
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

type int64Slice []int64

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SamplePercentile(values int64Slice, perc float64) float64 {
	ps := []float64{perc}

	scores := make([]float64, len(ps))
	size := len(values)
	if size > 0 {
		sort.Sort(values)
		for i, p := range ps {
			pos := p * float64(size+1) //ALTERNATIVELY, DROP THE +1
			if pos < 1.0 {
				scores[i] = float64(values[0])
			} else if pos >= float64(size) {
				scores[i] = float64(values[size-1])
			} else {
				lower := float64(values[int(pos)-1])
				upper := float64(values[int(pos)])
				scores[i] = lower + (pos-math.Floor(pos))*(upper-lower)
			}
		}
	}
	return scores[0]
}
