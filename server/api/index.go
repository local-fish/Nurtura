package handler

import (
	"log"
	"net/http"
	"sync"

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
			return
		}
		h = app.WithCORS(router)
	})
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	h.ServeHTTP(w, r)
}
