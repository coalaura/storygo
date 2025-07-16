package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/overview.txt
	PromptOverview string
)

func HandleOverview(w http.ResponseWriter, r *http.Request) {
	var overview GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&overview); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("overview: failed to decode request")
		log.WarningE(err)

		return
	}

	overview.Clean(true)

	request, err := CreateOverviewRequest(&overview)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("overview: failed to create request")
		log.WarningE(err)

		return
	}

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("overview: failed to start stream")
		log.WarningE(err)

		return
	}

	defer stream.Close()

	defer log.Debug("overview: finished overview")

	RespondWithStream(w, ctx, stream, "")
}

func CreateOverviewRequest(overview *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       "deepseek/deepseek-chat-v3-0324",
		Temperature: 0.8,
		MaxTokens:   2048,
		Stream:      true,
	}

	prompt, err := BuildOverviewPrompt(overview)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}

func BuildOverviewPrompt(overview *GenerationRequest) (string, error) {
	data := GenerationTemplate{
		Context:   overview.Context,
		Direction: overview.Direction,
	}

	if overview.Image != nil {
		description, err := GetImageDescription(*overview.Image)
		if err != nil {
			return "", err
		}

		data.Image = description
	}

	var prompt bytes.Buffer

	if err := OverviewTmpl.Execute(&prompt, data); err != nil {
		return "", err
	}

	return prompt.String(), nil
}
