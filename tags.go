package main

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/tags.txt
	PromptTags string
)

type TagList struct {
	Tags []string `json:"tags"`
}

func HandleTags(w http.ResponseWriter, r *http.Request) {
	log.Info("tags: new request")
	defer log.Info("tags: finished request")

	var tags GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("suggestion: failed to decode request")
		log.WarningE(err)

		return
	}

	tags.Clean(true)

	request, err := CreateTagsRequest(&tags)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("tags: failed to create request")
		log.WarningE(err)

		return
	}

	debugd("tags-request", &request)

	debugf("tags: running completion")

	ctx := r.Context()

	completion, err := OpenRouterRunCompletion(ctx, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("tags: failed to run completion")
		log.WarningE(err)

		return
	}

	debugd("tags-completion", &completion)

	log.Debug("tags: parsing completion")

	var list TagList

	err = json.Unmarshal([]byte(completion.Choices[0].Message.Content.Text), &list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("tags: failed to parse completion")
		log.WarningE(err)

		return
	}

	RespondWithJSON(w, http.StatusOK, list.Tags)
}

func CreateTagsRequest(tags *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       TagsModel,
		Temperature: 0.8,
		MaxTokens:   256,
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	prompt, err := BuildPrompt(nil, TagsTmpl, tags, nil)
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}
