package handler

import (
	"net/http"
	"sync"
	"log"
	"strings"
	"os"
	
	"github.com/rs/cors"

	"github.com/local-fish/Nurtura/server/app"
)

var (
	once sync.Once
	h    http.Handler
	err  error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		var router http.Handler
		router, err = app.New()
		if err != nil {
			log.Printf("app init: %v", err)
		}

		origins := []string{"http://localhost:5173"}
		if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
			origins = append(origins, strings.Split(extra, ",")...)
		}

		h = cors.New(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true, // only if you use cookies
		}).Handler(router)
	})
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	h.ServeHTTP(w, r)
}
