package db

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/snowflake/v2"
)

// File is a single file
type File struct {
	url string

	ID snowflake.ID

	Filename    string
	ContentType string

	Source      string
	Description string

	Data []byte
}

// URL ...
func (f File) URL() string {
	return f.url
}

// AddFile adds a file
func (db *Db) AddFile(filename, contentType string, data []byte) (f *File, err error) {
	f = &File{}
	err = pgxscan.Get(context.Background(), db.Pool, f, "insert into files (id, filename, content_type, data) values ($1, $2, $3, $4) returning *", db.Snowflake.Get(), filename, contentType, data)
	if err != nil {
		return nil, err
	}

	if db.Config.Bot.Website != "" {
		f.url = fmt.Sprintf("%vfile/%v/%v", db.Config.Bot.Website, f.ID, f.Filename)
	}

	return f, err
}

// File gets a file from the database
func (db *Db) File(id snowflake.ID) (f File, err error) {
	err = pgxscan.Get(context.Background(), db.Pool, &f, "select * from files where id = $1", id)
	if err != nil {
		return
	}

	if db.Config.Bot.Website != "" {
		f.url = fmt.Sprintf("%vfile/%v/%v", db.Config.Bot.Website, f.ID, f.Filename)
	}

	return
}

// Files gets all files
func (db *Db) Files() (f []File, err error) {
	err = pgxscan.Select(context.Background(), db.Pool, &f, "select id, filename, content_type, source, description from files order by filename asc")
	return
}

// FileName returns files with the given string in their name
func (db *Db) FileName(s string) (f []File, err error) {
	err = pgxscan.Select(context.Background(), db.Pool, &f, "select id, filename, content_type, source, description from files where position(lower($1) in lower(filename)) > 0 order by filename asc", s)
	return
}
