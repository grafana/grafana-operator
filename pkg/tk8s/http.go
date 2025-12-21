package tk8s

import (
	"fmt"
	"net/http"
)

// GetJSONmux returns an http.ServeMux configured to serve endpoints and payloads specified in data map
func GetJSONmux(t tHelper, data map[string]string) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()

	for endpoint, payload := range data {
		mux.HandleFunc(endpoint, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, payload)
		})
	}

	return mux
}
