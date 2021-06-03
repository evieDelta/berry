package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/termora/berry/db"
)

func (s *site) term(c echo.Context) (err error) {
	var t *db.Term

	terms, err := s.db.GetTerms(0)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	name := c.Param("term")
	name, err = url.PathUnescape(name)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	for _, i := range terms {
		if strings.EqualFold(i.Name, name) {
			t = i
			break
		}
	}

	if t == nil {
		return c.NoContent(http.StatusNotFound)
	}

	t.Description = strings.ReplaceAll(t.Description, "(##", "(/term/")
	t.Note = strings.ReplaceAll(t.Note, "(##", "(/term/")
	t.ContentWarnings = strings.ReplaceAll(t.ContentWarnings, "(##", "(/term/")

	return c.Render(http.StatusOK, "term.html", (&renderData{
		Conf: s.conf,
		Term: t,
	}).parse(c))
}
