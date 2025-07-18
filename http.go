package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

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

	ch := make(chan string)

	go func() {
		defer close(ch)

		var reasoning bool

		for {
			response, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				ch <- err.Error()

				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			choice := response.Choices[0]

			if choice.FinishReason == openrouter.FinishReasonContentFilter {
				ch <- "[stopped due to content_filter]"

				return
			}

			content := choice.Delta.Content

			if content != "" {
				ch <- content
			} else if choice.Delta.Reasoning != nil {
				if !reasoning {
					reasoning = true

					ch <- "\x00"
				}
			}
		}
	}()

	var finished bool
	for {
		select {
		case <-ctx.Done():
			log.Debug("generation: client closed connection")

			return
		case chunk, ok := <-ch:
			if !ok {
				return
			}

			if stop != "" {
				if index := strings.Index(chunk, stop); index != -1 {
					chunk = chunk[:index]

					finished = true
				}
			}

			w.Write([]byte(chunk))
			flusher.Flush()

			if finished {
				return
			}
		}
	}
}
