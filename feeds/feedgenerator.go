package feeds

import (
	"strconv"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/gorilla/feeds"
	"github.com/termora/berry/db"
)

const (
	title       = " update feed"
	description = "Plural and LGBTQ+ terminology database"
	author      = " contributors"
	email       = "contact@"
	copyright   = "/about/license"

	feedlength = 40
)

var created = time.Unix(1648289818, 0)

type Feeds struct {
	schema   string
	SiteName string
	BaseURL  string

	db *db.DB

	feed *feeds.Feed

	lastUpdated time.Time
	cachedRSS   string
	cachedAtom  string
	cachedJSON  string

	mu sync.RWMutex
}

func (f *Feeds) RSS() (string, error) {
	err := f.Update()
	if err != nil {
		return "", err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.cachedRSS, nil
}

func (f *Feeds) Atom() (string, error) {
	err := f.Update()
	if err != nil {
		return "", err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.cachedAtom, nil
}

func (f *Feeds) JSON() (string, error) {
	err := f.Update()
	if err != nil {
		return "", err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.cachedJSON, nil
}

func (f *Feeds) Update() error {
	if time.Now().Sub(f.lastUpdated) < 24*time.Hour {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	terms, err := f.db.TermsSince(time.Now().AddDate(0, 0, -feedlength))
	if err != nil {
		return errors.Wrap(err, "fetching term changelog")
	}

	items := make([]*feeds.Item, 0, len(terms))

	for _, term := range terms {
		desc := term.Description
		if term.Warning() {
			desc = "CW: " + term.ContentWarnings
		}

		items = append(items, &feeds.Item{
			Title:       term.Name,
			Link:        &feeds.Link{Href: f.schema + f.BaseURL + "/term/" + strconv.Itoa(term.ID)},
			Description: desc,
			Id:          "term:" + strconv.Itoa(term.ID),
			Updated:     term.LastModified,
			Created:     term.Created,
		})
	}

	f.feed.Items = items
	atom, err := f.feed.ToAtom()
	if err != nil {
		return errors.Wrap(err, "writing atom feed")
	}
	f.cachedAtom = atom

	rss, err := f.feed.ToRss()
	if err != nil {
		return errors.Wrap(err, "writing rss feed")
	}
	f.cachedRSS = rss

	json, err := f.feed.ToJSON()
	if err != nil {
		return errors.Wrap(err, "writing json feed")
	}
	f.cachedJSON = json

	return nil
}

func New(db *db.DB, urlSchema, baseURL string) *Feeds {
	feed := &feeds.Feed{
		Title:       title,
		Description: description,
		Link:        &feeds.Link{Href: urlSchema + baseURL},
		Created:     created,
		Author:      &feeds.Author{Name: baseURL + author, Email: email + baseURL},
		Copyright:   urlSchema + baseURL + copyright,
	}

	return &Feeds{
		SiteName: baseURL,
		BaseURL:  baseURL,
		schema:   urlSchema,
		db:       db,
		feed:     feed,
	}
}
