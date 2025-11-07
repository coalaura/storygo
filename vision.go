package main

import (
	"context"
	_ "embed"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	_ "image/jpeg"
	_ "image/png"

	"github.com/coalaura/lock"
	"github.com/corona10/goimagehash"
	"github.com/gen2brain/webp"
	"github.com/revrost/go-openrouter"

	"github.com/go-chi/chi/v5"
)

var (
	//go:embed prompts/vision.txt
	PromptVision string

	processing = lock.NewLockMap[string]()
)

func HandleImageUpload(w http.ResponseWriter, r *http.Request) {
	input, hash, err := ReceiveImage(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warnln("upload: failed to receive image")
		log.Warnln(err)

		return
	}

	details := r.FormValue("details")

	debugf("upload: describing image %q", hash)

	err = DescribeImage(r.Context(), hash, input, details)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Warnln("upload: failed to describe image")
		log.Warnln(err)

		return
	}

	debugf("upload: finished image %q", hash)

	RespondWithText(w, 200, hash)
}

func HandleImageHash(w http.ResponseWriter, r *http.Request) {
	_, hash, err := ReceiveImage(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		log.Warnln("hash: failed to receive image")
		log.Warnln(err)

		return
	}

	path := ImageWebpPath(hash)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)

		debugf("hash: image %q does not exist yet", hash)

		return
	}

	debugf("hash: image %q already exists", hash)

	RespondWithText(w, 200, hash)
}

func HandleImageServe(directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")

		if !IsHashValid(hash) {
			w.WriteHeader(http.StatusBadRequest)

			log.Warnln("image: invalid hash")

			return
		}

		file, err := os.OpenFile(filepath.Join(directory, hash+".webp"), os.O_RDONLY, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			log.Warnln("image: failed to open image")
			log.Warnln(err)

			return
		}

		defer file.Close()

		RespondWithImage(w, file)
	}
}

func ReceiveImage(r *http.Request) (image.Image, string, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, "", err
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		return nil, "", err
	}

	debugf("decoding image")

	input, _, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	debugf("upload: hashing image")

	perception, err := goimagehash.PerceptionHash(input)
	if err != nil {
		return nil, "", err
	}

	hash := fmt.Sprintf("%x", perception.GetHash())

	return input, hash, nil
}

func DescribeImage(ctx context.Context, hash string, img image.Image, details string) error {
	processing.Lock(hash)
	defer processing.Unlock(hash)

	path := ImageWebpPath(hash)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	if _, err := os.Stat("images"); os.IsNotExist(err) {
		os.Mkdir("images", 0755)
	}

	err := EncodeWebP(img, path)
	if err != nil {
		return err
	}

	uri, err := ReadImageAsDataURL(hash, VisionModelUseCompatibility)
	if err != nil {
		os.Remove(path)

		return err
	}

	var suffix string

	if details != "" {
		suffix = fmt.Sprintf("\n\nNotes: \"%s\"", details)
	}

	request := openrouter.ChatCompletionRequest{
		Model:       VisionModel,
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
								URL:    uri,
								Detail: openrouter.ImageURLDetailHigh,
							},
						},
					},
				},
			},
		},
	}

	debugd("vision-request", &request)

	debugf("upload: running completion")

	completion, err := OpenRouterRunCompletion(ctx, request)
	if err != nil {
		os.Remove(path)

		return err
	}

	debugd("vision-completion", &completion)

	path = ImageTextPath(hash)

	return os.WriteFile(path, []byte(completion.Choices[0].Message.Content.Text), 0644)
}

func EncodeWebP(img image.Image, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	err = webp.Encode(file, img, webp.Options{
		Quality: 95,
		Method:  4,
	})
	if err != nil {
		os.Remove(path)

		return err
	}

	return nil
}

func ImageWebpPath(hash string) string {
	return filepath.Join("images", hash+".webp")
}

func ImageTextPath(hash string) string {
	return filepath.Join("images", hash+".txt")
}

func IsHashValid(hash string) bool {
	rgx := regexp.MustCompile(`(?m)^[a-f0-9]+$`)

	return rgx.MatchString(hash)
}
