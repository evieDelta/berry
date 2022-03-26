package main

import (
	"context"
	"crypto/tls"
	"text/template"
	"time"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/Masterminds/sprig"
	"github.com/termora/berry/db"
	"github.com/termora/berry/db/search/typesense"
	"github.com/termora/berry/feeds"
	"go.uber.org/zap"
)

const mimeType = "text/gemini; lang=en"

type site struct {
	close context.CancelFunc

	db    *db.DB
	conf  conf
	sugar *zap.SugaredLogger

	gemini *gemini.Server
	mux    *gemini.Mux
	certs  func(hostname string) (*tls.Certificate, error)

	templ *template.Template

	feeds *feeds.Feeds
}

type conf struct {
	DatabaseURL string `yaml:"database_url"`

	BaseURL     string `yaml:"base_url"`
	FeedBaseURL string `yaml:"feed_base_url"`

	CertificatePath string `yaml:"certificate_path"`

	SiteName string `yaml:"site_name"`
	Invite   string `yaml:"invite_url"`
	Git      string
	Contact  bool

	Typesense struct {
		URL string
		Key string
	}
}

func (s *site) Run() {
	var err error

	var ctx context.Context
	ctx, s.close = context.WithCancel(context.Background())

	s.templ, err = template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(funcMap()).
		ParseFS(templFS, "templates/*.gotmpl")

	if err != nil {
		s.sugar.Fatal("Template Error:", err)
	}

	s.db, err = db.Init(s.conf.DatabaseURL, s.sugar)
	if err != nil {
		s.sugar.Fatalf("Error connecting to database: %v", err)
	}
	s.db.TermBaseURL = "/term/"
	s.sugar.Info("Connected to database")

	s.feeds = feeds.New(s.db, "gemini://", s.conf.FeedBaseURL)
	err = s.feeds.Update()
	if err != nil {
		s.sugar.Errorf("Error updating feeds: %v", err)
	}

	// Typesense requires a bot running to sync terms
	if s.conf.Typesense.URL != "" && s.conf.Typesense.Key != "" {
		s.db.Searcher, err = typesense.New(s.conf.Typesense.URL, s.conf.Typesense.Key, s.db.Pool, s.sugar.Debugf)
		if err != nil {
			s.sugar.Fatalf("Couldn't connect to Typesense: %v", err)
		}
		s.sugar.Info("Connected to Typesense server")
	}

	s.mux = &gemini.Mux{}
	s.gemini = &gemini.Server{
		Handler:        gemini.TimeoutHandler(s.mux, time.Second*15, "Server Action Timeout"), // chi's .Use is nicer for middlewares smh
		Addr:           s.conf.BaseURL,
		GetCertificate: s.certs,
	}

	s.mux.HandleFunc("/", s.index)
	s.mux.Handle("/tag/", gemini.StripPrefix("/tag/", gemini.HandlerFunc(s.tag)))
	s.mux.Handle("/term/", gemini.StripPrefix("/term/", gemini.HandlerFunc(s.term)))
	s.mux.HandleFunc("/search/", s.search)
	s.mux.Handle("/about/", gemini.StripPrefix("/about/", gemini.HandlerFunc(s.staticPage)))
	s.mux.Handle("/file/", gemini.StripPrefix("/file/", gemini.HandlerFunc(s.file)))
	// not currently used, can be uncommented if needed
	// s.mux.Handle("/static/", gemini.StripPrefix("/static/", gemini.FileServer(os.DirFS("static"))))

	s.mux.HandleFunc("/robots.txt", s.robotstxt)
	s.mux.HandleFunc("/favicon.txt", s.favicontxt)
	s.mux.HandleFunc("/favicon.png", s.faviconpng)
	s.mux.HandleFunc("/favicon.ico", s.faviconpng)

	s.mux.Handle("/feeds/", gemini.StripPrefix("/feeds/", gemini.HandlerFunc(s.contentfeeds)))

	go func() {
		if err := s.gemini.ListenAndServe(ctx); err != nil {
			s.sugar.Info("Shutting down server. Error?", err)
		}
	}()
}
