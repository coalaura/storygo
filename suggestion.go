package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/suggestion.txt
	PromptSuggestion string
)

func HandleSuggestion(w http.ResponseWriter, r *http.Request) {
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

	completion, err := OpenRouterRunCompletion(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("suggestion: failed to start stream")
		log.WarningE(err)

		return
	}

	defer log.Debug("suggestion: finished completion")

	cleaned := strings.ReplaceAll(completion.Choices[0].Message.Content.Text, "\n\n", "\n")

	RespondWithText(w, 200, cleaned)
}

func CreateSuggestionRequest(model *Model, suggestion *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       model.Slug,
		Temperature: 0.8,
		MaxTokens:   256,
	}

	model.SetReasoning(&request)

	prompt, err := BuildPrompt(model, SuggestionTmpl, suggestion)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	if suggestion.Image != nil && model.Vision {
		img, err := GetImageMessage(*suggestion.Image)
		if err != nil {
			return request, err
		}

		request.Messages = append(request.Messages, img)
	}

	return request, nil
}
