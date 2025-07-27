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
	log.Debugf("Starting...\r")

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := IndexTmpl.Execute(w, Models)

		if err != nil {
			RespondWithText(w, http.StatusInternalServerError, err.Error())
		}
	})

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))

	r.Get("/image/{hash}", HandleImageServe)

	r.Post("/upload", HandleImageUpload)
	r.Post("/suggest", HandleSuggestion)
	r.Post("/overview", HandleOverview)
	r.Post("/generate", HandleGeneration)

	log.Debug("Listening at http://localhost:3344/")
	http.ListenAndServe(":3344", r)
}
