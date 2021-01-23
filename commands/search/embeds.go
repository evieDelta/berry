package search

import (
	"fmt"
	"strings"

	"github.com/starshine-sys/berry/db"
	"github.com/diamondburned/arikawa/v2/discord"
)

var emoji = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"}

func searchResultEmbed(search string, page, total, totalTerms int, s []*db.Term) discord.Embed {
	var desc string
	for i, t := range s {
		h := t.Headline
		if !strings.HasPrefix(t.Description, h[:10]) {
			h = "..." + h
		}
		if !strings.HasSuffix(t.Description, h[len(h)-10:]) {
			h = h + "..."
		}
		name := t.Name
		if len(t.Aliases) > 0 {
			name += fmt.Sprintf(" (%v)", strings.Join(t.Aliases, ", "))
		}
		desc += fmt.Sprintf("%v **%v**\n%v\n\n", emoji[i], name, h)
	}

	v := []discord.EmbedField{{
		Name:  "Usage",
		Value: "Use ⬅️ ➡️ to navigate between pages and the numbers to choose a term.",
	}}
	if totalTerms <= 5 {
		v = nil
	}
	return discord.Embed{
		Title:       fmt.Sprintf("Search results for \"%v\"", search),
		Description: desc,
		Color:       db.EmbedColour,
		Fields:      v,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("Results: %v | Page %v/%v", totalTerms, page, total),
		},
	}
}
