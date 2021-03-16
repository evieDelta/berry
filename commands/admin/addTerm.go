package admin

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
	"github.com/termora/berry/db"
)

func (c *Admin) addTerm(ctx *bcr.Context) (err error) {
	if err = ctx.CheckMinArgs(1); err != nil {
		_, err = ctx.Send("Please provide a term name.", nil)
		return err
	}

	term := &db.Term{Name: ctx.RawArgs}
	ctx.AdditionalParams["term"] = term

	_, err = ctx.Sendf("Creating a term with the name `%v`. To cancel at any time, send `cancel`.\nPlease type the name of the category this term belongs to:", ctx.RawArgs)
	if err != nil {
		return err
	}

	// oh gods there's so much nesting there's no way we're gonna comment this
	// whoever sees this later, good luck, please burn this - Jake
	ctx.AddMessageHandler(ctx.Channel.ID, ctx.Author.ID, func(ctx *bcr.Context, m discord.Message) {
		if m.Content == "cancel" {
			ctx.Send("Term creation cancelled.", nil)
			return
		}
		cat, err := c.DB.CategoryID(m.Content)
		if err != nil {
			_, err = ctx.Send("Could not find that category, cancelled.", nil)
			return
		}
		if cat == 0 {
			return
		}

		t := ctx.AdditionalParams["term"].(*db.Term)
		t.Category = cat
		ctx.AdditionalParams["term"] = t
		_, err = ctx.Sendf("Category set to `%v` (ID %v). Please type the description:", m.Content, cat)
		if err != nil {
			return
		}

		ctx.AddMessageHandler(ctx.Channel.ID, ctx.Author.ID, func(ctx *bcr.Context, m discord.Message) {
			if m.Content == "cancel" {
				ctx.Send("Term creation cancelled.", nil)
				return
			}
			t := ctx.AdditionalParams["term"].(*db.Term)
			t.Description = m.Content
			if len(t.Description) > 1800 {
				_, err = ctx.Send("Description too long (maximum 1800 characters).", nil)
				return
			}
			ctx.AdditionalParams["term"] = t
			_, err := ctx.Send("Description set. Please type the source:", nil)
			if err != nil {
				return
			}

			ctx.AddMessageHandler(ctx.Channel.ID, ctx.Author.ID, func(ctx *bcr.Context, m discord.Message) {
				if m.Content == "cancel" {
					ctx.Send("Term creation cancelled.", nil)
					return
				}
				t := ctx.AdditionalParams["term"].(*db.Term)
				t.Source = m.Content
				ctx.AdditionalParams["term"] = t
				_, err := ctx.Send("Source set. Please type a *newline separated* list of aliases/synonyms, or \"none\" to set no aliases:", nil)
				if err != nil {
					return
				}

				ctx.AddMessageHandler(ctx.Channel.ID, ctx.Author.ID, func(ctx *bcr.Context, m discord.Message) {
					if m.Content == "cancel" {
						ctx.Send("Term creation cancelled.", nil)
						return
					}
					t := ctx.AdditionalParams["term"].(*db.Term)
					t.Aliases = strings.Split(m.Content, "\n")
					if m.Content == "none" {
						t.Aliases = []string{}
					}

					msg, err := ctx.Send("Term finished. React with ✅ to finish adding it, or with ❌ to cancel. Preview:", t.TermEmbed(""))
					if err != nil {
						return
					}

					if yes, timeout := ctx.YesNoHandler(*msg, ctx.Author.ID); !yes || timeout {
						ctx.Send("Cancelled.", nil)
						return
					}
					t, err = c.DB.AddTerm(t)
					if err != nil {
						c.DB.InternalError(ctx, err)
						return
					}
				})
			})
		})
	})

	return
}
