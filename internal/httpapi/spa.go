package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SPA serves static files from root and falls back to index.html.
func SPA(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/.") {
			full := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
			st, err := os.Stat(full)
			if err == nil && !st.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}
