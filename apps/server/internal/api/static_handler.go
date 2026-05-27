package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) staticHandler() http.Handler {
	fileServer := http.FileServer(http.Dir(s.config.WebStaticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/sfu/") {
			http.NotFound(w, r)
			return
		}

		requestPath := filepath.Clean(r.URL.Path)
		if requestPath == "." || requestPath == "/" {
			requestPath = "index.html"
		}

		fullPath := filepath.Join(s.config.WebStaticDir, requestPath)
		if _, err := os.Stat(fullPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.ServeFile(w, r, filepath.Join(s.config.WebStaticDir, "index.html"))
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
