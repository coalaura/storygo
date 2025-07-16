package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
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

	request, err := CreateSuggestionRequest(&suggestion)
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

	RespondWithText(w, 200, completion.Choices[0].Message.Content.Text)
}

func CreateSuggestionRequest(suggestion *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       "deepseek/deepseek-chat-v3-0324",
		Temperature: 0.8,
		MaxTokens:   256,
	}

	prompt, err := BuildSuggestionPrompt(suggestion)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	os.WriteFile("test.txt", []byte(prompt), 0755)

	return request, nil
}

func BuildSuggestionPrompt(suggestion *GenerationRequest) (string, error) {
	suggestion.Text = strings.TrimSpace(suggestion.Text)
	suggestion.Text = strings.ReplaceAll(suggestion.Text, "\r\n", "\n")

	data := GenerationTemplate{
		Context: suggestion.Context,
		Story:   suggestion.Text,
	}

	if suggestion.Image != nil {
		description, err := GetImageDescription(*suggestion.Image)
		if err != nil {
			return "", err
		}

		data.Image = description
	}

	var prompt bytes.Buffer

	if err := SuggestionTmpl.Execute(&prompt, data); err != nil {
		return "", err
	}

	return prompt.String(), nil
}
