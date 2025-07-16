package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type GenerationRequest struct {
	Text      string   `json:"text"`
	Context   string   `json:"context"`
	Direction string   `json:"direction"`
	Tags      []string `json:"tags"`
	Image     *string  `json:"image"`
}

type GenerationTemplate struct {
	Tags      string
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

func (g *GenerationRequest) TagList() string {
	if len(g.Tags) == 0 {
		return ""
	}

	return fmt.Sprintf("\"%s\"", strings.Join(g.Tags, "\", \""))
}

func BuildPrompt(tmpl *template.Template, request *GenerationRequest) (string, error) {
	data := GenerationTemplate{
		Tags:      request.TagList(),
		Context:   request.Context,
		Story:     request.Text,
		Direction: request.Direction,
		Empty:     request.Text == "",
	}

	if request.Image != nil {
		description, err := GetImageDescription(*request.Image)
		if err != nil {
			return "", err
		}

		data.Image = description
	}

	var prompt bytes.Buffer

	if err := tmpl.Execute(&prompt, data); err != nil {
		return "", err
	}

	return prompt.String(), nil
}

func CleanText(text string, trim bool) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	if trim {
		text = strings.TrimSpace(text)
	}

	return text
}
