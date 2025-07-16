package main

type Model struct {
	Slug   string
	Name   string
	Vision bool
	Reason bool
	Tags   []string
}

var (
	GenerationModels = []*Model{
		// Strongest open model with minimal restrictions and excellent creative output
		NewModel("deepseek/deepseek-chat-v3-0324", "DeepSeek V3 0324", false, false, []string{"unmoderated", "default", "fast"}),
		// Google's flagship model with superior reasoning and structured output
		NewModel("google/gemini-2.5-pro", "Gemini 2.5 Pro", true, true, []string{"creative", "structured", "reliable"}),
		// Current best reasoning model with excellent creative capabilities
		NewModel("openai/o3", "OpenAI o3", false, true, []string{"creative", "literary", "moderated"}),
		// Elon's latest flagship model - excellent at creative and reasoning tasks
		NewModel("xai/grok-4", "Grok 4", false, true, []string{"unmoderated", "creative", "fast"}),
		// Anthropic's flagship with superior literary style and safety
		NewModel("anthropic/claude-4-opus", "Claude 4 Opus", true, true, []string{"literary", "verbose", "moderated"}),
		// Balanced Claude model with great creative writing capabilities
		NewModel("anthropic/claude-4-sonnet", "Claude 4 Sonnet", true, true, []string{"creative", "structured", "moderated"}),
		// Strong unmoderated model excellent for character-driven stories
		NewModel("nousresearch/hermes-3-llama-3.1-405b", "Hermes 3 405B Instruct", false, false, []string{"unmoderated", "character-driven", "conversational"}),
		// China's breakout model with exceptional creative writing performance
		NewModel("moonshot/kimi-k2", "Kimi K2", false, false, []string{"unmoderated", "creative", "literary"}),
		// OpenAI's balanced flagship model
		NewModel("openai/gpt-4o", "GPT-4o", true, false, []string{"versatile", "conversational", "moderated"}),
		// Cost-effective with strong performance
		NewModel("meta-llama/llama-3.3-70b-instruct", "Llama 3.3 70B Instruct", false, false, []string{"unmoderated", "fast", "casual"}),
		// Fast and efficient Google model
		NewModel("google/gemini-2.5-flash", "Gemini 2.5 Flash", false, true, []string{"fast", "structured", "versatile"}),
		// Efficient OpenAI model for quick tasks
		NewModel("openai/gpt-4o-mini", "GPT-4o Mini", false, false, []string{"fast", "concise", "moderated"}),
	}

	VisionModels = []*Model{
		// Top unmoderated vision model with excellent image-to-text capabilities
		NewModel("qwen/qwen2.5-vl-32b-instruct", "Qwen2.5 VL 32B Instruct", true, false, []string{"unmoderated", "default", "descriptive"}),
		// Strong unmoderated vision model from Arcee AI
		NewModel("arcee-ai/spotlight", "Spotlight", true, false, []string{"unmoderated", "fast", "descriptive"}),
		// Another strong unmoderated vision option
		NewModel("mistralai/mistral-small-3.1-24b-instruct", "Mistral Small 3.1 24B", true, false, []string{"unmoderated", "multilingual", "fast"}),
		// Google's flagship vision model
		NewModel("google/gemini-2.5-pro", "Gemini 2.5 Pro", true, true, []string{"creative", "structured", "reliable"}),
		// Claude's premium vision model
		NewModel("anthropic/claude-4-opus", "Claude 4 Opus", true, true, []string{"literary", "verbose", "moderated"}),
		// Balanced Claude vision model
		NewModel("anthropic/claude-4-sonnet", "Claude 4 Sonnet", true, true, []string{"creative", "structured", "moderated"}),
		// OpenAI's flagship vision model
		NewModel("openai/gpt-4o", "GPT-4o", true, false, []string{"versatile", "conversational", "moderated"}),
		// Fast Google vision model
		NewModel("google/gemini-2.5-flash", "Gemini 2.5 Flash", true, true, []string{"fast", "structured", "versatile"}),
	}
)

func NewModel(slug, name string, vision, reason bool, tags []string) *Model {
	return &Model{
		Slug:   slug,
		Name:   name,
		Vision: vision,
		Reason: reason,
		Tags:   tags,
	}
}
