package e2e

import (
	"encoding/json"
)

// prettyJSON formats payload for golden comparisons.
func prettyJSON(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
