package main

import (
	"strings"
	"text/template"
)

type GenerationRequest struct {
	Text      string  `json:"text"`
	Context   string  `json:"context"`
	Direction string  `json:"direction"`
	Image     *string `json:"image"`
}

type GenerationTemplate struct {
	Context   string
	Direction string
	Story     string
	Image     string
	Empty     bool
}

var (
	GenerationTmpl *template.Template
	SuggestionTmpl *template.Template
	OverviewTmpl   *template.Template
)

func init() {
	gen, err := template.New("generation").Parse(PromptGeneration)
	log.MustPanic(err)

	GenerationTmpl = gen

	sug, err := template.New("suggestion").Parse(PromptSuggestion)
	log.MustPanic(err)

	SuggestionTmpl = sug

	ovr, err := template.New("overview").Parse(PromptOverview)
	log.MustPanic(err)

	OverviewTmpl = ovr
}

func (g *GenerationRequest) Clean(trim bool) {
	g.Context = CleanText(g.Context, true)
	g.Direction = CleanText(g.Direction, true)

	g.Text = CleanText(g.Text, trim)
}

func CleanText(text string, trim bool) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	if trim {
		text = strings.TrimSpace(text)
	}

	return text
}
