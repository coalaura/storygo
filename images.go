package main

import (
	_ "embed"

	"github.com/revrost/go-openrouter"
)

var (
	//go:embed prompts/images.txt
	PromptImages string
)

// TODO: generate prompt then call replicate

func CreateImagePromptRequest(model *Model, image *GenerationRequest) (openrouter.ChatCompletionRequest, error) {
	request := openrouter.ChatCompletionRequest{
		Model:       ImagePromptModel,
		Temperature: 0.15,
		MaxTokens:   8192 * 2,
	}

	prompt, err := BuildPrompt(&Model{}, ImagesTmpl, image, map[string]any{
		"Vision": model.Vision,
	})
	if err != nil {
		return request, err
	}

	request.Messages = []openrouter.ChatCompletionMessage{
		openrouter.SystemMessage(prompt),
	}

	return request, nil
}
