package main

import (
	"io"
	"net/http"
)

func RespondWithText(w http.ResponseWriter, code int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)

	w.Write([]byte(text))
}

func RespondWithImage(w http.ResponseWriter, file io.Reader) {
	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(200)

	io.Copy(w, file)
}
