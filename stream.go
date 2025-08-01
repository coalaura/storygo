package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/revrost/go-openrouter"
)

type Chunk struct {
	State string `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
	Text  string `json:"text,omitempty"`
}

func ErrChunk(err error) Chunk {
	return Chunk{Error: err.Error()}
}

func StateChunk(state string) Chunk {
	return Chunk{State: state}
}

func TextChunk(text string) Chunk {
	return Chunk{Text: text}
}

func ReceiveStream(stream *openrouter.ChatCompletionStream, stop string, rch chan Chunk) {
	defer close(rch)

	var dbg collector
	defer dbg.Print("stream: %q")

	var (
		reasoning bool
		receiving bool
		finished  bool
	)

	rch <- StateChunk("waiting")

	for {
		response, err := stream.Recv()
		if err != nil {
			if finished || errors.Is(err, io.EOF) {
				return
			}

			rch <- ErrChunk(err)

			return
		}

		if len(response.Choices) == 0 {
			continue
		}

		choice := response.Choices[0]

		if choice.FinishReason == openrouter.FinishReasonContentFilter {
			rch <- ErrChunk(errors.New("[stopped due to content_filter]"))

			return
		}

		content := choice.Delta.Content

		if content != "" {
			if stop != "" {
				if index := strings.Index(content, stop); index != -1 {
					content = content[:index]

					finished = true

					debugf("received stop word")
				}
			}

			if !receiving {
				receiving = true

				rch <- StateChunk("receiving")
			}

			dbg.Write(content)

			rch <- TextChunk(content)
		} else if choice.Delta.Reasoning != nil {
			if !reasoning {
				reasoning = true

				rch <- StateChunk("reasoning")
			}
		}

		if finished {
			return
		}
	}
}

func CreateResponseStream(w http.ResponseWriter) (chan Chunk, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("failed to create flusher")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var (
		ch      = make(chan Chunk)
		encoder = json.NewEncoder(w)
	)

	go func() {
		for {
			chunk, ok := <-ch
			if !ok {
				return
			}

			encoder.Encode(chunk)
			w.Write([]byte("\n\n"))

			flusher.Flush()
		}
	}()

	return ch, nil
}
