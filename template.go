package main

import "text/template"

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
