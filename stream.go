package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/revrost/go-openrouter"
)

type Chunk struct {
	State    string  `json:"state,omitempty"`
	Error    string  `json:"error,omitempty"`
	Text     string  `json:"text,omitempty"`
	Progress float64 `json:"progress,omitempty"`
}

type Stream struct {
	closed uint32
	wg     sync.WaitGroup
	rch    chan Chunk
	done   chan struct{}

	mx    sync.Mutex
	state string
}

func (s *Stream) Send(chunk Chunk) bool {
	if atomic.LoadUint32(&s.closed) == 1 {
		return false
	}

	select {
	case s.rch <- chunk:
		return true
	case <-s.done:
		return false
	}
}

func (s *Stream) SetState(state string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.state == state {
		return
	}

	s.state = state

	s.Send(StateChunk(state))
}

func (s *Stream) Close() {
	if !atomic.CompareAndSwapUint32(&s.closed, 0, 1) {
		return
	}

	close(s.done)

	s.wg.Wait()

	close(s.rch)
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

func ReceiveStream(stream *openrouter.ChatCompletionStream, stop string, rStream *Stream) {
	var dbg collector
	defer dbg.Print("stream: %q")

	var finished bool

	rStream.SetState("waiting")

	for {
		response, err := stream.Recv()
		if err != nil {
			if finished || errors.Is(err, io.EOF) {
				return
			}

			rStream.Send(ErrChunk(err))

			return
		}

		if len(response.Choices) == 0 {
			continue
		}

		choice := response.Choices[0]

		if choice.FinishReason == openrouter.FinishReasonContentFilter {
			rStream.Send(ErrChunk(errors.New("[stopped due to content_filter]")))

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

			rStream.SetState("receiving")

			dbg.Write(content)

			if !rStream.Send(TextChunk(content)) {
				return
			}
		} else if choice.Delta.Reasoning != nil {
			rStream.SetState("reasoning")
		}

		if finished {
			return
		}
	}
}

func CreateResponseStream(w http.ResponseWriter, ctx context.Context) (*Stream, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("failed to create flusher")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var (
		encoder = json.NewEncoder(w)
		stream  = &Stream{
			rch:  make(chan Chunk, 10),
			done: make(chan struct{}),
		}
	)

	stream.wg.Add(1)

	go func() {
		defer stream.wg.Done()

		for {
			select {
			case <-ctx.Done():
				debugf("cancelled response stream")

				return
			case chunk, ok := <-stream.rch:
				if !ok {
					debugf("closed response stream")

					return
				}

				if err := encoder.Encode(chunk); err != nil {
					log.WarningE(err)
					return
				}

				if _, err := w.Write([]byte("\n\n")); err == nil {
					flusher.Flush()
				} else {
					log.WarningE(err)
				}
			}
		}
	}()

	return stream, nil
}
