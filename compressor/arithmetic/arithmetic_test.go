package arithmetic

import (
	"testing"
	"sort"
	"math"
)

func round(num float64) int {
    return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
    output := math.Pow(10, float64(precision))
    return float64(round(num * output)) / output
}

func TestEncodeLoop(t *testing.T) {
	symFreqsWhole := map[byte]float64{'0': 0.05, '1': 0.05, '2': 0.5, '3': 0.4}
	keys := make(sortBytes, 0)
	for k, _ := range symFreqsWhole {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	input := []byte("2320")
	_, gotTop, gotBot := encodeLoop(false, keys, symFreqsWhole, input)
	precision := 3
	gotTop, gotBot = toFixed(gotTop, precision), toFixed(gotBot, precision)
    if gotTop != float64(0.425) || gotBot != float64(0.42)  {
        t.Errorf("encodeLoop(keys, symFreqsWhole, input) = %f, %f; want 0.425, 0.42", gotTop, gotBot)
    }
}