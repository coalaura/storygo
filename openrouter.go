package main

import (
	"context"
	"errors"
	"time"

	"github.com/revrost/go-openrouter"
)

func OpenRouterClient() *openrouter.Client {
	return openrouter.NewClient(OpenRouterToken)
}

func OpenRouterRunCompletion(request openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error) {
	client := OpenRouterClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

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
