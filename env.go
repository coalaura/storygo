package main

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	OpenRouterToken             string
	ReplicateToken              string
	VisionModel                 string
	ImagePromptModel            string
	TagsModel                   string
	VisionModelUseCompatibility bool

	Debug bool
)

func init() {
	log.MustFail(godotenv.Load())

	Debug = os.Getenv("DEBUG") == "true"

	if OpenRouterToken = os.Getenv("OPENROUTER_TOKEN"); OpenRouterToken == "" {
		log.MustFail(errors.New("missing openrouter token"))
	}

	if ReplicateToken = os.Getenv("REPLICATE_TOKEN"); ReplicateToken == "" {
		log.Warnln("No replicate key configured")
	}

	if VisionModel = os.Getenv("VISION_MODEL"); VisionModel == "" {
		VisionModel = "qwen/qwen2.5-vl-32b-instruct"
	}

	if ImagePromptModel = os.Getenv("IMAGE_PROMPT_MODEL"); ImagePromptModel == "" {
		ImagePromptModel = "deepseek/deepseek-chat-v3-0324"
	}

	if TagsModel = os.Getenv("TAGS_MODEL"); TagsModel == "" {
		TagsModel = "nousresearch/hermes-3-llama-3.1-405b"
	}

	VisionModelUseCompatibility = os.Getenv("VISION_MODEL_USE_COMPATIBILITY") == "true"

	log.Printf("Vision-Model: %s\n", VisionModel)
	log.Printf("Image-Model:  %s\n", ImagePromptModel)
	log.Printf("Vision-Comp.: %v\n", VisionModelUseCompatibility)

	debugf("Debug mode enabled")
}
