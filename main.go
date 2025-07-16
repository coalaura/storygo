package main

import (
	"net/http"
	"os/exec"
	"runtime"
	"time"

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

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/*", http.StripPrefix("/", fs))

	r.Get("/image/{hash}", HandleImageServe)
	r.Post("/upload", HandleImageUpload)
	r.Post("/suggest", HandleSuggestion)
	r.Post("/generate", HandleGeneration)

	time.AfterFunc(time.Second, open)

	log.Debug("Listening at http://localhost:3344/")
	http.ListenAndServe(":3344", r)
}

func open() {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	args = append(args, "http://localhost:3344/")

	exec.Command(cmd, args...).Start()
}
