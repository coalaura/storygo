package main

import (
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

		log.Warnln("overview: failed to decode request")
		log.Warnln(err)

		return
	}

	overview.Clean(true)

	model := GetModel(overview.Model)
	if model == nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warnln("overview: missing model")

		return
	}

	request, err := CreateOverviewRequest(model, &overview)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warnln("overview: failed to create request")
		log.Warnln(err)

		return
	}

	debugd("overview-request", &request)

	debugf("overview: starting completion stream")

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warnln("overview: failed to start stream")
		log.Warnln(err)

		return
	}

	defer stream.Close()

	defer log.Println("overview: finished overview")

	RespondWithStream(w, ctx, stream, "")
}

func CreateOverviewRequest(model *Model, overview *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       model.Slug,
		Temperature: 0.8,
		MaxTokens:   2048,
		Stream:      true,
	}

	model.SetReasoning(&request)

	prompt, err := BuildPrompt(model, OverviewTmpl, overview, nil)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	if overview.Image != nil && model.Vision {
		msg, err := ReadImageAsCompletionMessage(*overview.Image, model.UseCompatibility)
		if err != nil {
			return request, err
		}

		request.Messages = append(request.Messages, *msg)
	}

	return request, nil
}
