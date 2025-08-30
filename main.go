package main

import (
	"net/http"

	"github.com/coalaura/plain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/revrost/go-openrouter"
)

var log = plain.New(plain.WithDate(plain.RFC3339Local))

func init() {
	openrouter.DisableLogs()
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(log.Middleware())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := IndexTmpl.Execute(w, map[string]any{
			"models":    Models,
			"images":    ImageModels,
			"styles":    ImageStyles,
			"replicate": ReplicateToken != "",
		})

		if err != nil {
			RespondWithText(w, http.StatusInternalServerError, err.Error())
		}
	})

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))

	r.Get("/i/{hash}", HandleImageServe("images"))
	r.Get("/g/{hash}", HandleImageServe("generated"))

	r.Post("/image/hash", HandleImageHash)
	r.Post("/image/upload", HandleImageUpload)
	r.Post("/image/create/{style}", HandleImageGenerate)

	r.Post("/tags", HandleTags)
	r.Post("/context", HandleContext)
	r.Post("/suggest", HandleSuggestion)
	r.Post("/overview", HandleOverview)
	r.Post("/generate", HandleGeneration)

	log.Println("Listening at http://localhost:3344/")
	http.ListenAndServe(":3344", r)
}
