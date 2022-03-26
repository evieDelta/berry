package main

import (
	"context"
	"embed"
	"io"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"

	"git.sr.ht/~adnano/go-gemini"
)

//go:embed static/pages/*
var staticPages embed.FS

var pages = map[string]staticPage{}
var pageMu sync.RWMutex

type staticPage struct {
	Content string
	Format  string
}

// limit page names to 64 characters
var pageRegex = regexp.MustCompile(`^\w{1,64}$`)

func (s *site) staticPage(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	name, err := url.PathUnescape(r.URL.Path)
	if err != nil {
		s.sugar.Errorf("error decoding url: %v, %v", err, r.Conn().RemoteAddr())
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}
	name = strings.ToLower(
		strings.TrimSpace(
			strings.TrimSuffix(
				strings.TrimSuffix(name, ".gmi"), ".md",
			),
		),
	)

	data := &renderData{
		Conf: s.conf,
		Path: r.URL.Path,
	}

	pageMu.RLock()
	t, ok := pages[name]
	pageMu.RUnlock()

	if !ok {
		if !pageRegex.MatchString(name) {
			w.WriteHeader(gemini.StatusNotFound, "Page: `"+name+"` not found")
			return
		}

		b, err := staticPages.ReadFile(path.Join("static/pages/", name+".md"))
		c, er2 := staticPages.ReadFile(path.Join("static/pages/", name+".gmi"))
		if err != nil && er2 != nil {
			w.WriteHeader(gemini.StatusNotFound, "Page: `"+name+"` not found")
			return
		}
		if err != nil && er2 == nil {
			t.Content = string(c)
			t.Format = mimeType
		} else if err == nil && er2 == nil {
			s.sugar.Errorf("page `%v` has both a markdown and gemtext version", name)
			t.Content = string(c)
			t.Format = mimeType
		} else {
			t.Format = "text/markdown; lang=en"
			t.Content = string(b)
		}

		pageMu.Lock()
		pages[name] = t
		pageMu.Unlock()
	}

	data.MD = t.Content
	var page string

	if t.Format == mimeType {
		page, err = s.Render("static", data)
	} else {
		page = t.Content
	}

	if err != nil {
		s.sugar.Error("error fetching tags: ", err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Something went wrong")
		return
	}

	w.SetMediaType(t.Format)
	_, err = io.WriteString(w, page)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}
