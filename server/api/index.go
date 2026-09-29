package handler

import (
	"net/http"
	"sync"
	"log"

	"github.com/local-fish/Nurtura/server/app"
)

var (
	once sync.Once
	h    http.Handler
	err  error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		h, err = app.New()
		if err != nil {
			log.Printf("app init: %v", err)
		}
	})
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	h.ServeHTTP(w, r)
}
