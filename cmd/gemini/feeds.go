package main

import (
	"context"
	"io"

	"git.sr.ht/~adnano/go-gemini"
)

func (s *site) contentfeeds(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	var data, ctype string
	var err error
	switch r.URL.Path {
	case "feed.rss", "rss.xml":
		ctype = "application/rss+xml"
		data, err = s.feeds.RSS()
	case "feed.atom", "atom.xml":
		ctype = "application/atom+xml"
		data, err = s.feeds.Atom()
	case "feed.json":
		ctype = "application/feed+json"
		data, err = s.feeds.JSON()
	default:
		ctype = "text/gemini"
		data, err = s.Render("feeds", &renderData{
			Conf: s.conf,
		})
	}

	if err != nil {
		s.sugar.Errorf("Error getting/updating Feeds: %v", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "something went wrong")
	}

	w.SetMediaType(ctype)
	_, err = io.WriteString(w, data)
	if err != nil {
		s.sugar.Errorf("Error uploading feed: %v", err)
	}
}
