package main

import (
	"errors"
	"io"
	"strings"

	"github.com/revrost/go-openrouter"
)

func ReceiveStream(stream *openrouter.ChatCompletionStream, stop string) (chan string, chan error) {
	var (
		rch = make(chan string)
		ech = make(chan error)
	)

	go func() {
		defer close(rch)
		defer close(ech)

		var dbg collector
		defer dbg.Print("stream: %q")

		var (
			reasoning bool
			finished  bool
		)

		for {
			response, err := stream.Recv()
			if err != nil {
				if finished || errors.Is(err, io.EOF) {
					return
				}

				ech <- err

				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			choice := response.Choices[0]

			if choice.FinishReason == openrouter.FinishReasonContentFilter {
				ech <- errors.New("[stopped due to content_filter]")

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

				dbg.Write(content)

				rch <- content
			} else if choice.Delta.Reasoning != nil {
				if !reasoning {
					reasoning = true

					rch <- "\x00"
				}
			}

			if finished {
				return
			}
		}
	}()

	return rch, ech
}
