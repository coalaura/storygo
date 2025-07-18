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

	model := GetModel(generation.Model)
	if model == nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("generation: missing model")

		return
	}

	request, err := CreateGenerationRequest(model, &generation)
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

func CreateGenerationRequest(model *Model, generation *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       model.Slug,
		Temperature: 0.8,
		MaxTokens:   1024,
		Stop:        []string{"\n"},
		Stream:      true,
	}

	model.SetReasoning(&request)

	prompt, err := BuildPrompt(model, GenerationTmpl, generation)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	if generation.Image != nil && model.Vision {
		msg, err := ReadImageAsCompletionMessage(*generation.Image)
		if err != nil {
			return request, err
		}

		request.Messages = append(request.Messages, *msg)
	}

	return request, nil
}
