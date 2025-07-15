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

type GenerateRequest struct {
	Text      string  `json:"text"`
	Context   string  `json:"context"`
	Direction string  `json:"direction"`
	Image     *string `json:"image"`
}

type GenerateTemplate struct {
	Context   string
	Direction string
	Story     string
	Image     string
	Empty     bool
}

var (
	//go:embed prompts/generation.txt
	PromptGeneration string

	GenerationTemplate = ParseTemplateOrPanic(PromptGeneration)
)

func HandleGeneration(w http.ResponseWriter, r *http.Request) {
	var generate GenerateRequest

	if err := json.NewDecoder(r.Body).Decode(&generate); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("generate: failed to decode request")
		log.WarningE(err)

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generate: failed to create flusher")

		return
	}

	request, err := CreateGenerateRequest(&generate)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generate: failed to create request")
		log.WarningE(err)

		return
	}

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("generate: failed to start stream")
		log.WarningE(err)

		return
	}

	defer stream.Close()

	defer log.Debug("generate: finished generation")

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
			log.Debug("generate: client closed connection")

			return
		case chunk, ok := <-ch:
			if !ok {
				return
			}

			if index := strings.Index(chunk, "\n"); index != -1 {
				chunk = chunk[:index]
			}

			chunk = CleanChunk(chunk)

			w.Write([]byte(chunk))
			flusher.Flush()

			if newline {
				return
			}
		}
	}
}

func CreateGenerateRequest(generate *GenerateRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       "deepseek/deepseek-chat-v3-0324",
		Temperature: 0.8,
		MaxTokens:   1024,
		Stop:        []string{"\n"},
		Stream:      true,
	}

	prompt, err := BuildPrompt(generate)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}

func BuildPrompt(generate *GenerateRequest) (string, error) {
	generate.Text = strings.TrimSpace(generate.Text)
	generate.Text = strings.ReplaceAll(generate.Text, "\r\n", "\n")

	data := GenerateTemplate{
		Context:   generate.Context,
		Direction: generate.Direction,
		Story:     generate.Text + "\n\n",
		Empty:     generate.Text == "",
	}

	if generate.Image != nil {
		description, err := GetImageDescription(*generate.Image)
		if err != nil {
			return "", err
		}

		data.Image = description
	}

	var prompt bytes.Buffer

	if err := GenerationTemplate.Execute(&prompt, data); err != nil {
		return "", err
	}

	return prompt.String(), nil
}

func ParseTemplateOrPanic(text string) *template.Template {
	tmpl, err := template.New("system_prompt").Parse(PromptGeneration)
	log.MustPanic(err)

	return tmpl
}

func CleanChunk(chunk string) string {
	chunk = strings.ReplaceAll(chunk, "—\"", "-\"")

	chunk = strings.ReplaceAll(chunk, "—", ", ")

	return chunk
}
