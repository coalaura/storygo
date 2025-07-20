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
		// Excellent unmoderated model for creative writing, a great default.
		NewModel("deepseek/deepseek-chat-v3-0324", "DeepSeek V3 0324", false, false, []string{"unmoderated", "default"}),
		// The best choice for high-quality, literary prose and safe content.
		NewModel("anthropic/claude-4-opus", "Claude 4 Opus", false, true, []string{"literary", "moderated"}),
		// Google's flagship with vision, great for creative but logical stories.
		NewModel("google/gemini-2.5-pro", "Gemini 2.5 Pro", true, true, []string{"creative", "structured"}),
		// OpenAI's versatile model with vision, excels at conversational storytelling.
		NewModel("openai/gpt-4o", "GPT-4o", true, false, []string{"versatile", "conversational"}),
		// A massive open model, excels at character-driven stories and dialogue.
		NewModel("nousresearch/hermes-3-llama-3.1-405b", "Hermes 3 405B Instruct", false, false, []string{"unmoderated", "character-driven"}),
		// A top-tier open model with exceptional literary writing performance.
		NewModel("moonshotai/kimi-k2", "Kimi K2", false, false, []string{"unmoderated", "literary"}),
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
