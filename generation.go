package main

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/revrost/go-openrouter"
)

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

	generation.Clean(false)

	request, err := CreateGenerationRequest(&generation)
	if err != nil {
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

	RespondWithStream(w, ctx, stream, "\n")
}

func CreateGenerationRequest(generation *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       "deepseek/deepseek-chat-v3-0324",
		Temperature: 0.8,
		MaxTokens:   1024,
		Stop:        []string{"\n"},
		Stream:      true,
	}

	prompt, err := BuildPrompt(GenerationTmpl, generation)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}
