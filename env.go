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
	VisionModelUseCompatibility bool
)

func init() {
	log.MustPanic(godotenv.Load())

	if OpenRouterToken = os.Getenv("OPENROUTER_TOKEN"); OpenRouterToken == "" {
		log.Panic(errors.New("missing openrouter token"))
	}

	if ReplicateToken = os.Getenv("REPLICATE_TOKEN"); ReplicateToken == "" {
		log.Warning("No replicate key configured")
	}

	if VisionModel = os.Getenv("VISION_MODEL"); VisionModel == "" {
		VisionModel = "qwen/qwen2.5-vl-32b-instruct"
	}

	if ImagePromptModel = os.Getenv("IMAGE_PROMPT_MODEL"); ImagePromptModel == "" {
		ImagePromptModel = "nousresearch/hermes-3-llama-3.1-405b"
	}

	VisionModelUseCompatibility = os.Getenv("VISION_MODEL_USE_COMPATIBILITY") == "true"

	log.Debugf("Vision-Model: %s\n", VisionModel)
	log.Debugf("Image-Model:  %s\n", ImagePromptModel)
	log.Debugf("Vision-Comp.: %v\n", VisionModelUseCompatibility)
}
