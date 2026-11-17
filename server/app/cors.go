package app

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

// WithCORS wraps h with the browser CORS policy shared by both entrypoints.
// Defaults to the Vite dev server origin; CORS_ORIGINS adds comma-separated extras.
func WithCORS(h http.Handler) http.Handler {
	origins := []string{"http://localhost:5173"}
	if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
	}
	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // only if you use cookies
	}).Handler(h)
}
