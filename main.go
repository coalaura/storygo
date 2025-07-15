package main

import (
	"net/http"

	"github.com/coalaura/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var log = logger.New().DetectTerminal().WithOptions(logger.Options{
	NoLevel: true,
})

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	// r.Use(adapter.Middleware(log))

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))

	r.Get("/image/{hash}", HandleImageServe)
	r.Post("/upload", HandleImageUpload)
	r.Post("/generate", HandleGeneration)

	log.Debug("Listening at http://localhost:3344/")
	http.ListenAndServe(":3344", r)
}
