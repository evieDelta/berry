package main

import (
	"context"
	"io"

	"git.sr.ht/~adnano/go-gemini"
)

func (s *site) index(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	tags, err := s.db.Tags()
	if err != nil {
		s.sugar.Error("error fetching tags: ", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Something went wrong")
		return
	}

	page, err := s.Render("index", &renderData{
		Conf: s.conf,
		Tags: tags,
	})
	if err != nil {
		s.sugar.Error("error fetching tags: ", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Something went wrong")
		return
	}

	w.SetMediaType(mimeType)
	_, err = io.WriteString(w, page)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}
