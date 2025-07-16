package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/revrost/go-openrouter"
)

type GenerationRequest struct {
	Text      string  `json:"text"`
	Context   string  `json:"context"`
	Direction string  `json:"direction"`
	Image     *string `json:"image"`
}

type GenerationTemplate struct {
	Context   string
	Direction string
	Story     string
	Image     string
	Empty     bool
}

var (
	//go:embed prompts/generation.txt
	PromptGeneration string
)

func HandleGeneration(w http.ResponseWriter, r *http.Request) {
	var generation GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&generation); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("generation: failed to decode request")
		log.WarningE(err)

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generation: failed to create flusher")

		return
	}

	request, err := CreateGenerationRequest(&generation)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generation: failed to create request")
		log.WarningE(err)

		return
	}

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generation: failed to start stream")
		log.WarningE(err)

		return
	}

	defer stream.Close()

	defer log.Debug("generation: finished generation")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string)

	go func() {
		defer close(ch)

		for {
			response, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				ch <- err.Error()

				return
			}

			if len(response.Choices) > 0 {
				ch <- response.Choices[0].Delta.Content
			}
		}
	}()

	var newline bool

	for {
		select {
		case <-ctx.Done():
			log.Debug("generation: client closed connection")

			return
		case chunk, ok := <-ch:
			if !ok {
				return
			}

			if index := strings.Index(chunk, "\n"); index != -1 {
				chunk = chunk[:index]
			}

			w.Write([]byte(chunk))
			flusher.Flush()

			if newline {
				return
			}
		}
	}
}

func CreateGenerationRequest(generation *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       "deepseek/deepseek-chat-v3-0324",
		Temperature: 0.8,
		MaxTokens:   1024,
		Stop:        []string{"\n"},
		Stream:      true,
	}

	prompt, err := BuildGenerationPrompt(generation)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}

func BuildGenerationPrompt(generation *GenerationRequest) (string, error) {
	generation.Text = strings.ReplaceAll(generation.Text, "\r\n", "\n")

	data := GenerationTemplate{
		Context:   generation.Context,
		Direction: generation.Direction,
		Story:     generation.Text,
		Empty:     generation.Text == "",
	}

	if generation.Image != nil {
		description, err := GetImageDescription(*generation.Image)
		if err != nil {
			return "", err
		}

		data.Image = description
	}

	var prompt bytes.Buffer

	if err := GenerationTmpl.Execute(&prompt, data); err != nil {
		return "", err
	}

	return prompt.String(), nil
}

func ParseTemplateOrPanic(name, text string) *template.Template {
	tmpl, err := template.New(name).Parse(PromptGeneration)
	log.MustPanic(err)

	return tmpl
}
