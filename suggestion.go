package main

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/suggestion.txt
	PromptSuggestion string
)

func HandleSuggestion(w http.ResponseWriter, r *http.Request) {
	log.Info("suggestion: new request")
	defer log.Info("suggestion: finished request")

	var suggestion GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&suggestion); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("suggestion: failed to decode request")
		log.WarningE(err)

		return
	}

	suggestion.Clean(true)

	model := GetModel(suggestion.Model)
	if model == nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("suggestion: missing model")

		return
	}

	request, err := CreateSuggestionRequest(model, &suggestion)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("suggestion: failed to create request")
		log.WarningE(err)

		return
	}

	debugd("suggestion-request", &request)

	debugf("suggestion: starting completion stream")

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("suggestion: failed to start stream")
		log.WarningE(err)

		return
	}

	defer stream.Close()

	defer log.Debug("suggestion: finished completion")

	RespondWithStream(w, ctx, stream, "")
}

func CreateSuggestionRequest(model *Model, suggestion *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       model.Slug,
		Temperature: 0.8,
		MaxTokens:   256,
		Stream:      true,
	}

	model.SetReasoning(&request)

	prompt, err := BuildPrompt(nil, SuggestionTmpl, suggestion, nil)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}
