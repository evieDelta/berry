package admin

import (
	"strconv"

	"github.com/Starshine113/bcr"
	"github.com/Starshine113/berry/db"
	"github.com/Starshine113/berry/misc"
	"github.com/diamondburned/arikawa/v2/discord"
)

func (c *commands) setFlags(ctx *bcr.Context) (err error) {
	if err = ctx.CheckRequiredArgs(2); err != nil {
		_, err = ctx.Send("", &discord.Embed{
			Title: "Flags",
			Description: `The possible flags are:
		- 1: hidden from search
		- 2: hidden from random
		- 4: show a warning
		- 8: hide from lists (including the website)
		These can be combined by adding the numbers together.`,
			Color: db.EmbedColour,
		})
		return err
	}

	id, err := strconv.Atoi(ctx.Args[0])
	if err != nil {
		_, err = ctx.Sendf("Your input `%v` was not a number.", ctx.Args[0])
		return
	}

	flags, err := strconv.ParseInt(ctx.Args[1], 0, 0)
	if err != nil {
		_, err = ctx.Sendf("Your input `%v` was not a number.", ctx.Args[1])
		return
	}

	err = c.db.SetFlags(id, db.TermFlag(flags))
	if err != nil {
		_, err = ctx.Send(misc.InternalError, nil)
		return err
	}

	_, err = ctx.Sendf("Updated the flags for %v to %v.", id, flags)
	return
}
