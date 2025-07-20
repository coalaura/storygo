package main

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	OpenRouterToken string
	VisionModel     string
)

func init() {
	log.MustPanic(godotenv.Load())

	if OpenRouterToken = os.Getenv("OPENROUTER_TOKEN"); OpenRouterToken == "" {
		log.Panic(errors.New("missing openrouter token"))
	}

	if VisionModel = os.Getenv("VISION_MODEL"); VisionModel == "" {
		VisionModel = "qwen/qwen2.5-vl-32b-instruct"
	}

	log.Debugf("Vision-Model: %s\n", VisionModel)
}
