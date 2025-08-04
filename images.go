package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "embed"
	_ "image/jpeg"
	_ "image/png"

	"github.com/corona10/goimagehash"
	"github.com/go-chi/chi/v5"
	"github.com/replicate/replicate-go"
	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/images.txt
	PromptImages string
)

func HandleImageGenerate(w http.ResponseWriter, r *http.Request) {
	if ReplicateToken == "" {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("image: missing replicate token")

		return
	}

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

	ctx := r.Context()

	rStream, err := CreateResponseStream(w, ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("image: failed to create response stream")
		log.WarningE(err)

		return
	}

	defer rStream.Close()

	rStream.SetState("prompt")

	prompt, err := CreateImagePrompt(ctx, model, &image, style)
	if err != nil {
		log.Warning("image: failed to create prompt")
		log.WarningE(err)

		rStream.Send(ErrChunk(err))

		return
	}

	rStream.SetState("image")

	resultUrl, err := CreateImage(ctx, model, prompt)
	if err != nil {
		log.Warning("image: failed to create image")
		log.WarningE(err)

		rStream.Send(ErrChunk(err))

		return
	}

	rStream.SetState("save")

	rStream.Send(TextChunk(resultUrl))
}

func CreateImagePromptRequest(model *Model, image *GenerationRequest, style string) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       ImagePromptModel,
		Temperature: 0.15,
		MaxTokens:   8192 * 2,
	}

	prompt, err := BuildPrompt(nil, ImagesTmpl, image, map[string]any{
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

func CreateImagePrompt(ctx context.Context, model *Model, image *GenerationRequest, style string) (string, error) {
	request, err := CreateImagePromptRequest(model, image, style)
	if err != nil {
		return "", err
	}

	debugd("image-prompt-request", &request)

	debugf("image: running completion")

	completion, err := OpenRouterRunCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	debugd("image-prompt-completion", &completion)

	return completion.Choices[0].Message.Content.Text, nil
}

func CreateImage(ctx context.Context, model *Model, prompt string) (string, error) {
	input := replicate.PredictionInput{
		"prompt": prompt,
	}

	prediction, err := ReplicateRunPrediction(ctx, model.Slug, input)
	if err != nil {
		return "", err
	}

	url, err := ResolveReplicateURL(*prediction)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", err
	}

	perception, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", perception.GetHash())

	path := filepath.Join("generated", hash+".webp")

	if _, err := os.Stat("generated"); os.IsNotExist(err) {
		os.Mkdir("generated", 0755)
	}

	err = EncodeWebP(img, path)
	if err != nil {
		return "", err
	}

	return hash, nil
}
