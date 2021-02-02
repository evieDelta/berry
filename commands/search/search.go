package search

import (
	"time"

	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/bot"
)

type commands struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) (m string, list []*bcr.Command) {
	c := commands{Bot: bot}

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "search",
		Aliases: []string{"s"},

		Summary:     "Search for a term",
		Description: "Search for a term. Prefix your search with `!` to show the first result.",
		Usage:       "[--category/-c <category>] [--show-hidden/-h] <search term>",

		Blacklistable: true,

		Command: c.search,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "random",
		Aliases: []string{"r"},

		Summary: "Show a random term",

		Cooldown:      time.Second,
		Blacklistable: true,

		Command: c.random,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "explain",
		Aliases: []string{"e", "ex"},

		Summary: "Show a single explanation, or a list of all explanations",
		Usage:   "[explanation]",

		Cooldown:      time.Second,
		Blacklistable: false,

		Command: c.explanation,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "list",
		Summary: "List all terms",

		Cooldown:      time.Second,
		Blacklistable: true,
		Command:       c.list,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "post",
		Summary: "Post a single term",
		Usage:   "<term ID> [channel]",

		Cooldown:      time.Second,
		Blacklistable: true,
		Command:       c.term,
	}))

	// aliases
	ps := bot.Router.AddCommand(bot.Router.AliasMust(
		"plural", nil,
		[]string{"search"},
		bcr.DefaultArgTransformer("-c plurality", ""),
	))
	// we need to set these manually, the default description doesn't cut it
	ps.Summary = "Search for a plurality-related term"
	ps.Description = "Search for a term in the `plurality` category. Prefix your search with `!` to show the first result."
	ps.Usage = "<search term>"

	ls := bot.Router.AddCommand(bot.Router.AliasMust(
		"lgbt", []string{"lgbtq", "l"},
		[]string{"search"},
		bcr.DefaultArgTransformer("-c lgbtq+", ""),
	))
	// same as above
	ls.Summary = "Search for a LGBTQ+-related term"
	ls.Description = "Search for a term in the `LGBTQ+` category. Prefix your search with `!` to show the first result."
	ls.Usage = "<search term>"

	list = append(list, c.initExplanations(bot.Router)...)
	list = append(list, ps, ls)
	return "Search commands", list
}
