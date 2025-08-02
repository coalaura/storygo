package main

import (
	"context"
	"fmt"

	"github.com/replicate/replicate-go"
)

func ReplicateClient() (*replicate.Client, error) {
	return replicate.NewClient(replicate.WithToken(ReplicateToken))
}

func ReplicateRunPrediction(ctx context.Context, model string, input replicate.PredictionInput) (*replicate.PredictionOutput, error) {
	client, err := ReplicateClient()
	if err != nil {
		return nil, err
	}

	output, err := client.Run(ctx, model, input, nil)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func ResolveReplicateURL(prediction replicate.PredictionOutput) (string, error) {
	urls, ok := prediction.([]interface{})
	if ok {
		prediction = urls[0]
	}

	url, ok := prediction.(string)
	if ok {
		return url, nil
	}

	return "", fmt.Errorf("expected replicate url, got %T: %#v", prediction, prediction)
}
