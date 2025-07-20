package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/jpeg"
	"os"

	"github.com/gen2brain/webp"
	"github.com/revrost/go-openrouter"
)

func ReadImageWebpData(hash string) ([]byte, error) {
	if !IsHashValid(hash) {
		return nil, errors.New("invalid hash")
	}

	return os.ReadFile(ImageWebpPath(hash))
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

func ReadImageAsDataURL(hash string, useCompatibility bool) (string, error) {
	data, err := ReadImageWebpData(hash)
	if err != nil {
		return "", err
	}

	mime := "webp"

	if useCompatibility {
		img, err := webp.Decode(bytes.NewReader(data))
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer

		err = jpeg.Encode(&buf, img, &jpeg.Options{
			Quality: 90,
		})
		if err != nil {
			return "", err
		}

		mime = "jpeg"
		data = buf.Bytes()
	}

	b64 := base64.StdEncoding.EncodeToString(data)

	return fmt.Sprintf("data:image/%s;base64,%s", mime, b64), nil
}

func ReadImageAsCompletionMessage(hash string, useCompatibility bool) (*openrouter.ChatCompletionMessage, error) {
	uri, err := ReadImageAsDataURL(hash, useCompatibility)
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
