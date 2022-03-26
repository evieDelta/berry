package main

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/starshine-sys/snowflake/v2"
)

func (s *site) file(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	urlraw, err := url.PathUnescape(r.URL.Path)
	if err != nil {
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}
	url := strings.SplitN(urlraw, "/", 2)

	i, err := strconv.ParseUint(url[0], 0, 0)
	if err != nil {
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}

	f, err := s.db.File(snowflake.ID(i))
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			w.WriteHeader(gemini.StatusNotFound, "File: `"+url[0]+"/"+url[1]+"` not found")
			return
		}

		s.sugar.Errorf("Error getting file %v: %v", i, err)
		w.WriteHeader(gemini.StatusTemporaryFailure, "Something went wrong")
		return
	}

	w.SetMediaType(f.ContentType)
	_, err = w.Write(f.Data)
	if err != nil {
		s.sugar.Error("error uploading", err)
	}
}
