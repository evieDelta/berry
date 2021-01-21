package admin

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"

	"github.com/Starshine113/bcr"
	"github.com/Starshine113/berry/db"
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/utils/sendpart"
)

func (c *commands) cmdGuilds(ctx *bcr.Context) (err error) {
	b := make([]string, 0)

	for _, g := range c.guilds {
		b = append(b, fmt.Sprintf(
			"Name = %v\nID = %v", g.Name, g.ID,
		))
	}
	s := strings.Join(b, "\n###\n")

	if len(s) <= 2000 {
		_, err = ctx.Send("", &discord.Embed{
			Title:       fmt.Sprintf("Guilds (%v)", len(c.guilds)),
			Description: "```ini\n" + s + "\n```",
			Color:       db.EmbedColour,
		})
		return err
	}

	fn := "guilds.txt"
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)
	zw.Name = fn
	_, err = zw.Write([]byte(s))
	if err != nil {
		return c.db.InternalError(ctx, err)
	}
	err = zw.Close()
	if err != nil {
		return c.db.InternalError(ctx, err)
	}
	fn += ".gz"

	file := sendpart.File{
		Name:   fn,
		Reader: buf,
	}

	_, err = ctx.Session.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         "Here you go!",
		Files:           []sendpart.File{file},
		AllowedMentions: ctx.Router.DefaultMentions,
	})
	return err
}
