package main

import (
	"context"
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

func RespondWithStream(w http.ResponseWriter, ctx context.Context, stream *openrouter.ChatCompletionStream, stop string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generation: failed to create flusher")

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rch, ech := ReceiveStream(stream, stop)

	for {
		select {
		case <-ctx.Done():
			log.Debug("generation: client closed connection")

			return
		case err, ok := <-ech:
			if !ok {
				return
			}

			w.Write([]byte(err.Error()))
			flusher.Flush()
		case chunk, ok := <-rch:
			if !ok {
				return
			}

			w.Write([]byte(chunk))
			flusher.Flush()
		}
	}
}
