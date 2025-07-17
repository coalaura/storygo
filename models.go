package main

import (
	"encoding/json"
	"net/http"
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
}

var (
	Models = []*Model{
		// Strongest open model with minimal restrictions and excellent creative output
		NewModel("deepseek/deepseek-chat-v3-0324", "DeepSeek V3 0324", false, false, []string{"unmoderated", "default"}),
		// Google's flagship model with superior reasoning and structured output
		NewModel("google/gemini-2.5-pro", "Gemini 2.5 Pro", true, true, []string{"creative", "structured"}),
		// Current best reasoning model with excellent creative capabilities
		NewModel("openai/o3", "OpenAI o3", false, true, []string{"literary", "moderated"}),
		// Elon's latest flagship model - excellent at creative and reasoning tasks
		NewModel("x-ai/grok-4", "Grok 4", false, true, []string{"unmoderated", "creative"}),
		// Anthropic's flagship with superior literary style and safety
		NewModel("anthropic/claude-opus-4", "Claude 4 Opus", false, true, []string{"literary", "verbose"}),
		// Balanced Claude model with great creative writing capabilities
		NewModel("anthropic/claude-sonnet-4", "Claude 4 Sonnet", false, true, []string{"creative", "structured"}),
		// Strong unmoderated model excellent for character-driven stories
		NewModel("nousresearch/hermes-3-llama-3.1-405b", "Hermes 3 405B Instruct", false, false, []string{"unmoderated", "character-driven"}),
		// China's breakout model with exceptional creative writing performance
		NewModel("moonshotai/kimi-k2", "Kimi K2", false, false, []string{"unmoderated", "literary"}),
		// OpenAI's balanced flagship model
		NewModel("openai/gpt-4o", "GPT-4o", true, false, []string{"versatile", "conversational"}),
		// Cost-effective with strong performance
		NewModel("meta-llama/llama-3.3-70b-instruct", "Llama 3.3 70B Instruct", false, false, []string{"unmoderated", "casual"}),
		// Fast and efficient Google model
		NewModel("google/gemini-2.5-flash", "Gemini 2.5 Flash", false, true, []string{"fast", "versatile"}),
		// Efficient OpenAI model for quick tasks
		NewModel("openai/gpt-4o-mini", "GPT-4o Mini", false, false, []string{"fast", "concise"}),
	}

	ModelList, _ = json.Marshal(Models)
)

func (m *Model) Path(path string) string {
	slug := m.Slug

	if index := strings.Index(slug, "/"); index != -1 {
		slug = slug[index+1:]
	}

	return filepath.Join("images", path)
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

func NewModel(slug, name string, vision, reason bool, tags []string) *Model {
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

func HandleModelsServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(ModelList)
}
