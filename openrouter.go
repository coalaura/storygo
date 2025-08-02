package main

import (
	"context"
	"errors"

	"github.com/revrost/go-openrouter"
)

func OpenRouterClient() *openrouter.Client {
	return openrouter.NewClient(OpenRouterToken)
}

func OpenRouterAdjustRequest(request openrouter.ChatCompletionRequest) openrouter.ChatCompletionRequest {
	/*
		request.Provider = &openrouter.ChatProvider{
			Quantizations: []string{
				"fp8", "fp16", "bf16", "fp32", "unknown",
			},
		}
	*/

	return request
}

func OpenRouterRunCompletion(ctx context.Context, request openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error) {
	client := OpenRouterClient()

	request = OpenRouterAdjustRequest(request)

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

	request = OpenRouterAdjustRequest(request)

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}

	return stream, nil
}
