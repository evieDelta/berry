package main

import (
	"context"
	"io"

	"git.sr.ht/~adnano/go-gemini"
)

func (s *site) search(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	q, err := gemini.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		s.sugar.Errorf("error decoding url: %v, %v", err, r.Conn().RemoteAddr())
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}

	if q == "" {
		w.WriteHeader(gemini.StatusInput, "Input Search Query")
		return
	}

	//	q := template.HTML(bluemonday.UGCPolicy().Sanitize(c.QueryParam("q")))
	terms, err := s.db.Search(q, 0, []string{})

	var page string
	if err != nil || len(terms) == 0 {
		page, err = s.Render("no-results", &renderData{
			Conf:  s.conf,
			Path:  r.URL.Path,
			Query: q,
		})
	} else {
		page, err = s.Render("search-results", &renderData{
			Conf:  s.conf,
			Terms: terms,
			Path:  r.URL.Path,
			Query: q,
		})
	}

	if err != nil {
		s.sugar.Error("error performing search: ", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Something went wrong")
		return
	}

	w.SetMediaType(mimeType)
	_, err = io.WriteString(w, page)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}
