package main

import (
	"log"
	"net/http"
	"time"
	"os"

	"github.com/local-fish/Nurtura/server/app"

	"github.com/joho/godotenv"
)

// This is for Local Testing without Vercel.
func main() {
	_ = godotenv.Load(".env")
	h, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr: ":" + port,
		// Same CORS policy as the Vercel handler, so the Vite dev server can
		// talk to this directly instead of going through `vercel dev`.
		Handler:           app.WithCORS(h),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
