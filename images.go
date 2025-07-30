package main

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/images.txt
	PromptImages string
)

// TODO: call replicate and generate image

func HandleImageGenerate(w http.ResponseWriter, r *http.Request) {
	log.Info("image: new request")
	defer log.Info("image: finished request")

	index, err := strconv.Atoi(chi.URLParam(r, "style"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: invalid style index")
		log.WarningE(err)

		return
	}

	if index < 0 || index >= len(ImageStyles) {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: style index too high/low")

		return
	}

	style := ImageStyles[index]

	debugf("image: style %q", style)

	var image GenerationRequest

	if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: failed to decode request")
		log.WarningE(err)

		return
	}

	image.Clean(true)

	model := GetImageModel(image.Model)
	if model == nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: missing model")

		return
	}

	prompt, err := CreateImagePrompt(model, &image, style)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("image: failed to create prompt")
		log.WarningE(err)

		return
	}

	RespondWithText(w, 200, prompt)
}

func CreateImagePromptRequest(model *Model, image *GenerationRequest, style string) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       ImagePromptModel,
		Temperature: 0.15,
		MaxTokens:   8192 * 2,
	}

	prompt, err := BuildPrompt(&Model{}, ImagesTmpl, image, map[string]any{
		"Vision": model.Vision,
		"Style":  style,
	})
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}

func CreateImagePrompt(model *Model, image *GenerationRequest, style string) (string, error) {
	request, err := CreateImagePromptRequest(model, image, style)
	if err != nil {
		return "", err
	}

	debugd("image-prompt-request", &request)

	debugf("image: running completion")

	completion, err := OpenRouterRunCompletion(request)
	if err != nil {
		return "", err
	}

	debugd("image-prompt-completion", &completion)

	return completion.Choices[0].Message.Content.Text, nil
}
