package main

import (
	"context"
	"errors"

	"github.com/revrost/go-openrouter"
)

func OpenRouterClient() *openrouter.Client {
	return openrouter.NewClient(OpenRouterToken, openrouter.WithXTitle("StoryGo"), openrouter.WithHTTPReferer("https://github.com/coalaura/storygo"))
}

func OpenRouterRunCompletion(ctx context.Context, request openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error) {
	client := OpenRouterClient()

	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, errors.New("received no choices")
	}

	if response.Choices[0].FinishReason == "error" {
		return nil, errors.New("response finished with error")
	}

	return &response, nil
}

func OpenRouterStartStream(ctx context.Context, request openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionStream, error) {
	client := OpenRouterClient()

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}

	return stream, nil
}
