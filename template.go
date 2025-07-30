package main

import (
	"bytes"
	"fmt"
	html "html/template"
	"strings"
	"text/template"
)

type GenerationRequest struct {
	Model     string   `json:"model"`
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

	Extra map[string]any
}

var (
	IndexTmpl *html.Template

	GenerationTmpl *template.Template
	SuggestionTmpl *template.Template
	OverviewTmpl   *template.Template
	ImagesTmpl     *template.Template
)

func init() {
	IndexTmpl = html.Must(html.ParseFiles("static/index.html"))

	gen, err := template.New("generation").Parse(PromptGeneration)
	log.MustPanic(err)

	GenerationTmpl = gen

	sug, err := template.New("suggestion").Parse(PromptSuggestion)
	log.MustPanic(err)

	SuggestionTmpl = sug

	ovr, err := template.New("overview").Parse(PromptOverview)
	log.MustPanic(err)

	OverviewTmpl = ovr

	img, err := template.New("images").Parse(PromptImages)
	log.MustPanic(err)

	ImagesTmpl = img
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

func BuildPrompt(model *Model, tmpl *template.Template, request *GenerationRequest, extra map[string]any) (string, error) {
	data := GenerationTemplate{
		Tags:      request.TagList(),
		Context:   request.Context,
		Story:     request.Text,
		Direction: request.Direction,
		Empty:     request.Text == "",
		Extra:     extra,
	}

	if request.Image != nil && !model.Vision {
		description, err := ReadImageTextData(*request.Image)
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
