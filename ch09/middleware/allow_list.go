package middleware

import (
	"net/http"
	"path"
	"path/filepath"
)

func AllowListMiddleware(allows []string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanGivenPath := path.Clean(r.URL.Path)

		for _, p := range allows {
			cleanPath := filepath.Clean(p)
			if eq, _ := filepath.Match(cleanPath, cleanGivenPath); eq {
				h.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	})
}
