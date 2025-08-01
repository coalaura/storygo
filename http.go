package main

import (
	"io"
	"net/http"

	"github.com/revrost/go-openrouter"
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

func RespondWithStream(w http.ResponseWriter, stream *openrouter.ChatCompletionStream, stop string) {
	response, err := CreateResponseStream(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.WarningE(err)

		return
	}

	ReceiveStream(stream, stop, response)
}
