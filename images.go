package main

import "strings"

type ImageModel struct {
	Key      string   `json:"key"`
	Slug     string   `json:"slug"`
	Name     string   `json:"name"`
	Vision   bool     `json:"vision"`
	Tags     []string `json:"tags"`
	Strength string
	Weakness string
}

var (
	// Curated image-generation models
	ImageModels = []*ImageModel{
		// Google's newest photoreal SOTA; superb quality, but heavily moderated
		NewImageModel("google/imagen-4-ultra", "Imagen-4 Ultra", false, []string{"moderated", "quality"}),
		// OpenAI's flagship multimodal creator; great at following complex prompts & edits
		NewImageModel("openai/gpt-image-1", "GPT-Image-1", false, []string{"creative", "structured"}),
		// Black-Forest Labs 4-MP monster; "Ultra" mode for max detail, Raw for realism, unmoderated
		NewImageModel("black-forest-labs/flux-1.1-pro-ultra", "Flux 1.1 Pro Ultra", false, []string{"unmoderated", "quality"}),
		// Fast 4-step SD3.5 distill; good balance of speed and realism
		NewImageModel("stability-ai/stable-diffusion-3.5-large-turbo", "SD-3.5 Large Turbo", false, []string{"fast", "versatile"}),
		// ByteDance's bilingual model; excels at layout & short text in images
		NewImageModel("bytedance/seedream-3", "Seedream 3", false, []string{"creative", "structured"}),
		// Flux "schnell" = 1-4-step speed demon; great for cheap drafts, unmoderated
		NewImageModel("black-forest-labs/flux-schnell", "Flux Schnell", false, []string{"fast", "unmoderated"}),
		// Ideogram's turbo tier; best in class for legible embedded text
		NewImageModel("ideogram-ai/ideogram-v3-turbo", "Ideogram v3 Turbo", false, []string{"structured", "fast"}),
		// Recraft V3 for vector / logo / SVG-style outputs
		NewImageModel("recraft-ai/recraft-v3", "Recraft V3", false, []string{"creative", "structured"}),
	}

	ImageStyles = []string{
		"Photorealism",
		"Anime",
		"Graphic Design",
		"Painterly",
		"Concept Art",
	}
)

func NewImageModel(slug, name string, vision bool, tags []string) *ImageModel {
	key := slug

	if i := strings.Index(key, "/"); i != -1 {
		key = key[i+1:]
	}

	return &ImageModel{
		Key:    key,
		Slug:   slug,
		Name:   name,
		Vision: vision,
		Tags:   tags,
	}
}
