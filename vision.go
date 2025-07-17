package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	_ "image/jpeg"
	_ "image/png"

	"github.com/coalaura/lock"
	"github.com/corona10/goimagehash"
	"github.com/gen2brain/webp"
	"github.com/nfnt/resize"
	"github.com/revrost/go-openrouter"

	"github.com/go-chi/chi/v5"
)

var (
	//go:embed prompts/vision.txt
	PromptVision string

	processing = lock.NewLockMap[string]()
)

func HandleImageUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("upload: failed to parse multipart form")
		log.WarningE(err)

		return
	}

	details := r.FormValue("details")

	file, _, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("upload: failed to read form file")
		log.WarningE(err)

		return
	}

	log.Debug("upload: decoding image")

	input, _, err := image.Decode(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("upload: failed to decode image")
		log.WarningE(err)

		return
	}

	input = resize.Thumbnail(1024, 1024, input, resize.Lanczos3)

	log.Debug("upload: hashing image")

	perception, err := goimagehash.PerceptionHash(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("upload: failed to hash image")
		log.WarningE(err)

		return
	}

	hash := fmt.Sprintf("%x", perception.GetHash())

	log.Debugf("upload: describing image (%s)\n", hash)

	err = DescribeImage(hash, input, details)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warning("upload: failed to describe image")
		log.WarningE(err)

		return
	}

	log.Debugf("upload: finished image (%s)\n", hash)

	RespondWithText(w, 200, hash)
}

func HandleImageServe(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	if !IsHashValid(hash) {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: invalid hash")

		return
	}

	file, err := os.OpenFile(ImageWebPPath(hash), os.O_RDONLY, 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warning("image: failed to open image")
		log.WarningE(err)

		return
	}

	defer file.Close()

	RespondWithImage(w, file)
}

func DescribeImage(hash string, img image.Image, details string) error {
	processing.Lock(hash)
	defer processing.Unlock(hash)

	path := ImageWebPPath(hash)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	if _, err := os.Stat("images"); os.IsNotExist(err) {
		os.Mkdir("images", 0755)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	var buf bytes.Buffer

	writer := io.MultiWriter(file, &buf)

	err = webp.Encode(writer, img, webp.Options{
		Quality: 95,
		Method:  4,
	})
	if err != nil {
		return err
	}

	var suffix string

	if details != "" {
		suffix = fmt.Sprintf("\n\nNotes: \"%s\"", details)
	}

	request := openrouter.ChatCompletionRequest{
		Model:       "qwen/qwen2.5-vl-32b-instruct",
		Temperature: 0.15,
		MaxTokens:   8192 * 2,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage(PromptVision),
			{
				Role: openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{
					Multi: []openrouter.ChatMessagePart{
						{
							Type: openrouter.ChatMessagePartTypeText,
							Text: "Analyze this image and generate the (as detailed as possible) description based on the system instructions." + suffix,
						},
						{
							Type: openrouter.ChatMessagePartTypeImageURL,
							ImageURL: &openrouter.ChatMessageImageURL{
								URL:    "data:image/webp;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
								Detail: openrouter.ImageURLDetailHigh,
							},
						},
					},
				},
			},
		},
	}

	completion, err := OpenRouterRunCompletion(request)
	if err != nil {
		return err
	}

	path = ImageTextPath(hash)

	return os.WriteFile(path, []byte(completion.Choices[0].Message.Content.Text), 0644)
}

func ImageWebPPath(hash string) string {
	return filepath.Join("images", hash+".webp")
}

func ImageTextPath(hash string) string {
	return filepath.Join("images", hash+".txt")
}

func IsHashValid(hash string) bool {
	rgx := regexp.MustCompile(`(?m)^[a-f0-9]+$`)

	return rgx.MatchString(hash)
}

func GetImageDescription(hash string) (string, error) {
	if !IsHashValid(hash) {
		return "", errors.New("invalid hash")
	}

	data, err := os.ReadFile(ImageTextPath(hash))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func GetImageMessage(hash string) (openrouter.ChatCompletionMessage, error) {
	message := openrouter.ChatCompletionMessage{
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
						URL:    "data:image/webp;base64,",
						Detail: openrouter.ImageURLDetailAuto,
					},
				},
			},
		},
	}

	if !IsHashValid(hash) {
		return message, errors.New("invalid hash")
	}

	data, err := os.ReadFile(ImageWebPPath(hash))
	if err != nil {
		return message, err
	}

	message.Content.Multi[1].ImageURL.URL += base64.StdEncoding.EncodeToString(data)

	return message, nil
}
