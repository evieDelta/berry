package main

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

func (s *site) search(c echo.Context) (err error) {
	q := template.HTML(bluemonday.UGCPolicy().Sanitize(c.QueryParam("q")))
	terms, err := s.db.Search(c.QueryParam("q"), 0, []string{})

	if err != nil || len(terms) == 0 {
		return c.Render(http.StatusNotFound, "noQuery.html", renderData{
			Conf:  s.conf,
			Query: q,
		})
	}

	return c.Render(http.StatusOK, "results.html", renderData{
		Conf:  s.conf,
		Terms: terms,
		Query: q,
	})
}
