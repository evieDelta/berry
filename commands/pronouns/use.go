package pronouns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/db"
)

func (c *commands) use(ctx *bcr.Context) (err error) {
	if err = ctx.CheckMinArgs(1); err != nil {
		return c.list(ctx)
	}

	sets, err := c.DB.GetPronoun(strings.Split(ctx.Args[0], "/")...)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			_, err = ctx.Sendf("Couldn't find any pronoun sets from your input. Try `%vlist-pronouns` for a list of all pronouns; if it's not on there, feel free to submit it with `%vsubmit-pronouns`!", ctx.Router.Prefixes[0], ctx.Router.Prefixes[0])
			return
		}
		if err == db.ErrTooManyForms {
			_, err = ctx.Sendf("You gave too many forms! Input up to five forms, separated with a slash (`/`).")
			return err
		}
		return c.DB.InternalError(ctx, err)
	}

	if len(sets) > 1 {
		s := fmt.Sprintf("Found more than one set matching your input! Please be more specific.\nSets found:\n")
		for _, p := range sets {
			s += fmt.Sprintf("- %s\n", p)
		}
		_, err = ctx.NewMessage().Content(s).BlockMentions().Send()
		return err
	}
	// use the first set
	set := sets[0]

	if tmplCount == 0 {
		_, err = ctx.Send("There are no examples available for pronouns! If you think this is in error, please join the bot support server and ask there.", nil)
		return err
	}

	var (
		b strings.Builder
		e = make([]discord.Embed, 0)
	)

	e = append(e, discord.Embed{
		Title:       fmt.Sprintf("%v/%v pronouns", set.Subjective, set.Objective),
		Description: fmt.Sprintf("**%s**\n\nTo see these pronouns in action, use the arrow reactions on this message!", set),
		Color:       db.EmbedColour,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v | Page 1/%v", set.ID, tmplCount+1),
		},
	})

	useSet := &db.PronounSet{
		Subjective: set.Subjective,
		Objective:  set.Objective,
		PossDet:    set.PossDet,
		PossPro:    set.PossPro,
		Reflexive:  set.Reflexive,
	}
	if len(ctx.Args) > 1 {
		useSet.Subjective = ctx.Args[1]
	}

	for i := 0; i < tmplCount; i++ {
		err = templates.ExecuteTemplate(&b, strconv.Itoa(i), useSet)
		if err != nil {
			return c.DB.InternalError(ctx, err)
		}
		e = append(e, discord.Embed{
			Title:       fmt.Sprintf("%v/%v pronouns", set.Subjective, set.Objective),
			Description: b.String(),
			Color:       db.EmbedColour,
			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("ID: %v | Page %v/%v", set.ID, i+2, tmplCount+1),
			},
		})
		b.Reset()
	}

	_, err = ctx.PagedEmbed(e, false)
	return
}
