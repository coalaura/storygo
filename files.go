package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/revrost/go-openrouter"
)

func ReadImageJpegData(hash string) ([]byte, error) {
	if !IsHashValid(hash) {
		return nil, errors.New("invalid hash")
	}

	return os.ReadFile(ImageJpegPath(hash))
}

func ReadImageTextData(hash string) (string, error) {
	if !IsHashValid(hash) {
		return "", errors.New("invalid hash")
	}

	data, err := os.ReadFile(ImageTextPath(hash))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func ReadImageAsDataURL(hash string) (string, error) {
	data, err := ReadImageJpegData(hash)
	if err != nil {
		return "", err
	}

	b64 := base64.RawStdEncoding.EncodeToString(data)

	return fmt.Sprintf("data:image/jpeg;base64,%s", b64), nil
}

func ReadImageAsCompletionMessage(hash string) (*openrouter.ChatCompletionMessage, error) {
	uri, err := ReadImageAsDataURL(hash)
	if err != nil {
		return nil, err
	}

	return &openrouter.ChatCompletionMessage{
		Role: openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: "Here is a key image for the story. It could be a character's appearance, a specific location, or a pivotal scene. Use this image as a guiding reference for atmosphere and consistency, letting its details subtly inform your response.",
				},
				{
					Type: openrouter.ChatMessagePartTypeImageURL,
					ImageURL: &openrouter.ChatMessageImageURL{
						URL:    uri,
						Detail: openrouter.ImageURLDetailAuto,
					},
				},
			},
		},
	}, nil
}
