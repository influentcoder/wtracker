package main

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/influentcoder/wtracker/internal/bitcoin"
	"github.com/influentcoder/wtracker/internal/blockchain"
	"github.com/influentcoder/wtracker/internal/cache"
	"github.com/influentcoder/wtracker/internal/config"
	"github.com/influentcoder/wtracker/internal/handlers"
	"github.com/influentcoder/wtracker/ui"
)

func main() {
	cfg := config.Load()

	clients := map[string]blockchain.Client{
		"bitcoin": bitcoin.NewClient(),
	}

	c := cache.New()
	api := handlers.NewAPI(cfg, clients, c)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/whales", api.ListWhales)
		r.Get("/whales/{address}", api.GetWhale)
		r.Get("/price", api.GetPrice)
	})

	// Serve embedded frontend
	staticRoot, err := fs.Sub(ui.StaticFS, "static")
	if err != nil {
		log.Fatal("embed sub:", err)
	}
	staticServer := http.FileServer(http.FS(staticRoot))

	// /static/* serves CSS, JS, etc.
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", staticServer).ServeHTTP(w, r)
	})

	// / serves index.html
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staticServer.ServeHTTP(w, r)
	})

	addr := ":" + cfg.Port
	log.Printf("wtracker listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
