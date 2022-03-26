package main

import (
	"context"
	"io"
	"net/url"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/termora/berry/db"
)

func (s *site) tag(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	tag, err := url.PathUnescape(r.URL.Path)
	if err != nil {
		s.sugar.Errorf("error decoding url: %v, %v", err, r.Conn().RemoteAddr())
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}

	var terms []*db.Term
	if tag == "untagged" || tag == "" {
		terms, err = s.db.UntaggedTerms()
	} else {
		terms, err = s.db.TagTerms(tag)
	}
	if err != nil {
		s.sugar.Errorf("error fetching tag content: %v", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Database Error")
		return
	}

	page, err := s.Render("terms", &renderData{
		Conf:  s.conf,
		Tag:   tag,
		Terms: terms,
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
