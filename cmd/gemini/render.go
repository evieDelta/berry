package main

import (
	"embed"
	"strings"

	"github.com/termora/berry/db"
)

//go:embed templates/*
var templFS embed.FS

// Render a page
func (s *site) Render(name string, data interface{}) (string, error) {
	w := &strings.Builder{}
	err := s.templ.ExecuteTemplate(w, name, data)

	return w.String(), err
}

type renderData struct {
	Conf  conf
	Path  string
	Tag   string
	Tags  []string
	Term  *db.Term
	Terms []*db.Term

	TermLinks TermLinks

	Query string
	MD    string
}

type TermLinks struct {
	ContentWarning []linkPair
	Description    []linkPair
	Note           []linkPair
	Source         []linkPair
}
