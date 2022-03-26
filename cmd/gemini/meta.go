package main

import (
	"context"
	_ "embed"
	"io"

	"git.sr.ht/~adnano/go-gemini"
)

func (s *site) robotstxt(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	w.SetMediaType("text/plain")
	_, err := io.WriteString(w, robotstxt)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}

func (s *site) favicontxt(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	w.SetMediaType("text/plain")
	_, err := io.WriteString(w, favicontxt)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}

func (s *site) faviconpng(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	w.SetMediaType("image/png")
	_, err := w.Write(faviconpng)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}

const robotstxt = `User-agent: *
Disallow: /file
Disallow: /search
Disallow: /static`

// yes i seen the drama, please do not bother us with any purity evangelizing
// we respectfully, do not care - evie
const favicontxt = "📕" // unicode book

//go:embed static/favicon.png
var faviconpng []byte
