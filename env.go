package main

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	OpenRouterToken             string
	VisionModel                 string
	VisionModelUseCompatibility bool
)

func init() {
	log.MustPanic(godotenv.Load())

	if OpenRouterToken = os.Getenv("OPENROUTER_TOKEN"); OpenRouterToken == "" {
		log.Panic(errors.New("missing openrouter token"))
	}

	if VisionModel = os.Getenv("VISION_MODEL"); VisionModel == "" {
		VisionModel = "qwen/qwen2.5-vl-32b-instruct"
	}

	VisionModelUseCompatibility = os.Getenv("VISION_MODEL_USE_COMPATIBILITY") == "true"

	log.Debugf("Vision-Model: %s\n", VisionModel)
	log.Debugf("Vision-Comp.: %v\n", VisionModelUseCompatibility)
}
