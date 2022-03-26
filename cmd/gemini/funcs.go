package main

import (
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// HeadlineLen is the length of a headline (used in embeds)
const HeadlineLen = 250

func funcMap() template.FuncMap {
	return map[string]interface{}{
		"ne": func(a, b interface{}) bool {
			return a != b
		},

		"urlEncode": func(s string) string {
			return url.PathEscape(s)
		},
		"urlDecode": func(s string) string {
			u, err := url.PathUnescape(s)
			if err != nil {
				return ""
			}
			return u
		},
		"timeToDate": func(t time.Time) string {
			return t.Format("2006-01-02") + " / " + relativeTime(t)
		},

		"isPlural": func(i int) bool { return i != 1 },
		"title":    strings.Title,

		"quoteMultiline": func(s string) string {
			if strings.Count(s, "\n") < 1 {
				return s
			}

			return strings.ReplaceAll(s, "\n", "\n> ")
		},

		"headline": func(in string) string {
			if len(in) <= HeadlineLen {
				return in
			}

			slice := strings.Split(in, " ")
			buf := slice[0]
			for _, s := range slice[1:] {
				if len(buf) > HeadlineLen {
					buf += "..."
					break
				}
				buf += " " + s
			}
			return strings.TrimSpace(buf)
		},
	}
}

func relativeTime(t time.Time) string {
	now := time.Now()

	dist := now.Sub(t)
	daydist := now.YearDay() - t.YearDay()

	switch {
	case dist < -time.Minute:
		return (-dist).Round(time.Second).String() + " in the future"
	case dist < time.Second:
		return "Now"
	case dist < time.Minute:
		i := int64(dist.Seconds())
		if i == 1 {
			return "a second ago"
		}
		return strconv.FormatInt(i, 10) + " seconds ago"
	case dist < time.Hour:
		i := int64(dist.Minutes())
		if i == 1 {
			return "a minute ago"
		}
		return strconv.FormatInt(i, 10) + " minutes ago"
	case dist < time.Hour*24:
		i := int64(dist.Hours())
		if i == 1 {
			return "an hour ago"
		}
		return strconv.FormatInt(i, 10) + " hours ago"
	case dist < time.Hour*48:
		return "yesterday"
	case daydist < 7 && now.Year() == t.Year():
		return "last " + t.Weekday().String()
	case daydist < 28 && now.Year() == t.Year():
		i := daydist / 7
		if i == 1 {
			return "last week"
		}
		return strconv.Itoa(i) + " weeks ago"
	case daydist < 332:
		return "last " + t.Month().String()
	default:
		i := now.Year() - t.Year()
		if i == 1 {
			return "last year"
		}
		return strconv.Itoa(i) + " years ago"
	}
}
