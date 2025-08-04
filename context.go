package main

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/context.txt
	PromptContext string
)

func HandleContext(w http.ResponseWriter, r *http.Request) {
	var context GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&context); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("context: failed to decode request")
		log.WarningE(err)

		return
	}

	context.Clean(true)

	model := GetModel(context.Model)
	if model == nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("context: missing model")

		return
	}

	request, err := CreateContextRequest(model, &context)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("context: failed to create request")
		log.WarningE(err)

		return
	}

	debugd("context-request", &request)

	debugf("context: starting completion stream")

	ctx := r.Context()

	stream, err := OpenRouterStartStream(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("context: failed to start stream")
		log.WarningE(err)

		return
	}

	defer stream.Close()

	defer log.Debug("context: finished completion")

	RespondWithStream(w, ctx, stream, "")
}

func CreateContextRequest(model *Model, context *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       model.Slug,
		Temperature: 0.8,
		MaxTokens:   512,
		Stream:      true,
	}

	model.SetReasoning(&request)

	prompt, err := BuildPrompt(nil, ContextTmpl, context, nil)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}
