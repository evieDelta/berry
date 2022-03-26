package main

import (
	"context"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/termora/berry/db"
)

var numberRegex = regexp.MustCompile(`^\d+$`)

func (s *site) term(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	var t *db.Term

	name := r.URL.Path
	name, err := url.PathUnescape(name)
	if err != nil {
		s.sugar.Errorf("error decoding url: %v, %v", err, r.Conn().RemoteAddr())
		w.WriteHeader(gemini.StatusBadRequest, "Invalid Input")
		return
	}

	if numberRegex.MatchString(name) {
		id, _ := strconv.Atoi(name)

		t, err = s.db.GetTerm(id)
		if err != nil {
			s.sugar.Errorf("error fetching tag content: %v", err)
			w.WriteHeader(gemini.StatusTemporaryFailure, "Database Error")
			return
		}
	} else {
		terms, err := s.db.GetTerms(0)
		if err != nil {
			s.sugar.Errorf("error fetching tag content: %v", err)
			w.WriteHeader(gemini.StatusTemporaryFailure, "Database Error")
			return
		}

		for _, i := range terms {
			if strings.EqualFold(i.Name, name) {
				t = i
				break
			}
		}
	}

	if t == nil {
		w.WriteHeader(gemini.StatusNotFound, "Term not found")
		return
	}

	cw, cwlinks := linkReformatter(s.db.LinkTerms(t.ContentWarnings))
	t.ContentWarnings = cw
	desc, desclinks := linkReformatter(s.db.LinkTerms(t.Description))
	t.Description = desc
	note, notelinks := linkReformatter(s.db.LinkTerms(t.Note))
	t.Note = note
	source, sourcelinks := linkReformatter(t.Source)
	t.Source = source
	if t.Disputed() {
		t.Note = strings.TrimSpace(t.Note + "\n\n" + db.DisputedText)
	}

	page, err := s.Render("term", &renderData{
		Conf: s.conf,
		Term: t,

		TermLinks: TermLinks{
			ContentWarning: cwlinks,
			Description:    desclinks,
			Note:           notelinks,
			Source:         sourcelinks,
		},
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

var linkRegex = regexp.MustCompile(`\[([^\]\[\n]+)\]\((\/term\/(?:\d+))\)|https?:\/\/(\S+)`)

type linkPair struct {
	Name string
	Dest string
}

// reformat the already formatted links to work with gemtext lol
func linkReformatter(s string) (out string, links []linkPair) {
	parsed := linkRegex.FindAllStringSubmatch(s, -1)
	if len(parsed) == 0 {
		return s, nil
	}

	for i, link := range parsed {
		if link[3] != "" {
			parsed[i][1] = link[3]
			parsed[i][2] = "https://" + link[3]
		}
	}

	links = make([]linkPair, 0, len(parsed))
	for i, link := range parsed {
		links = append(links, linkPair{
			Name: strconv.Itoa(i+1) + ". " + link[1],
			Dest: link[2],
		})
	}

	var i int
	s = linkRegex.ReplaceAllStringFunc(s, func(s string) string {
		i++
		return "^" + parsed[i-1][1] + "[" + strconv.Itoa(i) + "]"
	})

	return s, links
}
