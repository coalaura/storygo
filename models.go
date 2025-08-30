package main

import (
	"path/filepath"
	"strings"

	"github.com/revrost/go-openrouter"
)

type Model struct {
	Key    string   `json:"key"`
	Slug   string   `json:"slug"`
	Name   string   `json:"name"`
	Vision bool     `json:"vision"`
	Reason bool     `json:"reason"`
	Tags   []string `json:"tags"`

	UseCompatibility bool `json:"-"`
}

var (
	Models = []*Model{
		// Excellent unmoderated default for creative prose; great balance of quality and speed.
		NewModel("deepseek/deepseek-chat-v3-0324", "DeepSeek V3 0324", false, false, []string{"unmoderated", "default"}),
		// Anthropic's current sweet spot for long-form literary style; vision + long context.
		NewModel("anthropic/claude-sonnet-4", "Claude 4 Sonnet", true, true, []string{"literary", "moderated"}),
		// OpenAI's newest flagship; strong global coherence, planning, and style control.
		NewModel("openai/gpt-5", "GPT-5", false, true, []string{"versatile", "creative"}),
		// Google's multimodal reasoning work-horse; structured output and tool use.
		NewModel("google/gemini-2.5-pro", "Gemini 2.5 Pro", true, true, []string{"creative", "structured"}),
		// xAI's vision model; witty personality, fewer policy limits, excellent creativity.
		NewModel("x-ai/grok-4", "Grok 4", true, true, []string{"unmoderated", "creative"}, true),
		// OpenAI's conversational all‑rounder with fast vision; great pacing and dialogue.
		NewModel("openai/gpt-4o", "GPT-4o", true, false, []string{"versatile", "conversational"}),
		// Mid‑tier Claude; cheaper than 4‑series, still strong at analysis and consistency.
		NewModel("anthropic/claude-sonnet-3.7", "Claude 3.7 Sonnet", true, true, []string{"creative", "structured"}),
		// Meta's new open MoE; very long context + vision; solid general writing.
		NewModel("meta-llama/llama-4-maverick", "Llama-4 Maverick", true, true, []string{"unmoderated", "versatile"}),
		// Qwen3 instruction model with strong adherence and long context; good value.
		NewModel("qwen/Qwen3-30B-A3B-Instruct-2507", "Qwen3 30B Instruct", false, true, []string{"creative", "structured"}),
		// Zhipu's fast multimodal model; good for quick drafts and agentic flows.
		NewModel("z-ai/glm-4.5-air", "GLM-4.5 Air", true, true, []string{"fast", "versatile"}),
		// Massive open model; shines in dialogue, role‑play, and character consistency.
		NewModel("nousresearch/hermes-4-405b", "Hermes 4 405B", false, false, []string{"unmoderated", "character-driven"}),
		// Open‑weight Chinese model; excels at descriptive, imaginative prose.
		NewModel("moonshotai/kimi-k2", "Kimi K2", false, false, []string{"unmoderated", "literary"}),
	}

	ImageModels = []*Model{
		// Google's newest photoreal SOTA; superb quality, but heavily moderated
		NewModel("google/imagen-4-ultra", "Imagen-4 Ultra", false, false, []string{"moderated", "quality"}),
		// OpenAI's flagship multimodal creator; great at following complex prompts & edits
		NewModel("openai/gpt-image-1", "GPT-Image-1", false, false, []string{"creative", "structured"}),
		// Black-Forest Labs 4-MP monster; "Ultra" mode for max detail, Raw for realism, unmoderated
		NewModel("black-forest-labs/flux-1.1-pro-ultra", "Flux 1.1 Pro Ultra", false, false, []string{"unmoderated", "quality"}),
		// Fast 4-step SD3.5 distill; good balance of speed and realism
		NewModel("stability-ai/stable-diffusion-3.5-large-turbo", "SD-3.5 Large Turbo", false, false, []string{"fast", "versatile"}),
		// ByteDance's bilingual model; excels at layout & short text in images
		NewModel("bytedance/seedream-3", "Seedream 3", false, false, []string{"creative", "structured"}),
		// Flux "schnell" = 1-4-step speed demon; great for cheap drafts, unmoderated
		NewModel("black-forest-labs/flux-schnell", "Flux Schnell", false, false, []string{"fast", "unmoderated"}),
		// Ideogram's turbo tier; best in class for legible embedded text
		NewModel("ideogram-ai/ideogram-v3-turbo", "Ideogram v3 Turbo", false, false, []string{"structured", "fast"}),
		// Recraft V3 for vector / logo / SVG-style outputs
		NewModel("recraft-ai/recraft-v3", "Recraft V3", false, false, []string{"creative", "structured"}),
	}

	ImageStyles = []string{
		"Photorealism",
		"Anime",
		"Graphic Design",
		"Painterly",
		"Concept Art",
	}
)

func (m *Model) Path(path string) string {
	slug := m.Slug

	if index := strings.Index(slug, "/"); index != -1 {
		slug = slug[index+1:]
	}

	return filepath.Join("images", slug)
}

func (m *Model) SetReasoning(request *openrouter.ChatCompletionRequest) {
	if !m.Reason {
		return
	}

	thinking := 256

	request.Reasoning = &openrouter.ChatCompletionReasoning{
		MaxTokens: &thinking,
	}
}

func NewModel(slug, name string, vision, reason bool, tags []string, useCompatibility ...bool) *Model {
	key := slug

	if index := strings.Index(key, "/"); index != -1 {
		key = key[index+1:]
	}

	return &Model{
		Key:    key,
		Slug:   slug,
		Name:   name,
		Vision: vision,
		Reason: reason,
		Tags:   tags,

		UseCompatibility: len(useCompatibility) > 0 && useCompatibility[0],
	}
}

func GetModel(key string) *Model {
	for _, model := range Models {
		if model.Key == key {
			return model
		}
	}

	return nil
}

func GetImageModel(key string) *Model {
	for _, model := range ImageModels {
		if model.Key == key {
			return model
		}
	}

	return nil
}
