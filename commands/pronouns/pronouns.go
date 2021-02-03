package pronouns

import (
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/bot"
)

var templates = template.Must(template.New("").Funcs(funcs()).ParseGlob("commands/pronouns/examples/*"))
var tmplCount int

// initialise number of templates
func init() {
	files, err := ioutil.ReadDir("commands/pronouns/examples")
	if err != nil {
		panic(err)
	}
	tmplCount = len(files)
}

type commands struct {
	*bot.Bot

	submitCooldown *ttlcache.Cache
}

// Init ...
func Init(bot *bot.Bot) (m string, list []*bcr.Command) {
	c := &commands{
		Bot:            bot,
		submitCooldown: ttlcache.NewCache(),
	}
	c.submitCooldown.SkipTTLExtensionOnHit(true)

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "pronouns",
		Aliases: []string{"pronoun", "neopronoun", "neopronouns"},

		Summary: "Show pronouns (with optional name) used in a sentence",
		Usage:   "<pronouns> [name]",

		Blacklistable: true,
		Cooldown:      time.Second,
		Command:       c.use,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "list-pronouns",
		Aliases: []string{"pronoun-list", "listpronouns", "pronounlist"},

		Summary: "Show a list of all pronouns",

		Blacklistable: true,
		Cooldown:      time.Second,
		Command:       c.list,
	}))

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name: "submit-pronouns",

		Summary: "Submit a pronoun set",
		Usage:   "<pronouns, forms separated with />",

		Blacklistable: true,
		Command:       c.submit,
	}))

	return "Pronoun commands", list
}

func funcs() map[string]interface{} {
	return map[string]interface{}{
		"title": strings.Title,
	}
}
