package admin

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Starshine113/crouter"
	"github.com/Starshine113/flagparser"
	"github.com/Starshine113/termbot/db"
	"github.com/bwmarrin/discordgo"
)

type e struct {
	ExportDate time.Time  `json:"export_date"`
	Terms      []*db.Term `json:"terms"`
}

func (c *commands) export(ctx *crouter.Ctx) (err error) {
	export := e{ExportDate: time.Now().UTC()}

	fp, _ := flagparser.NewFlagParser(flagparser.Bool("gzip", "gz"), flagparser.String("out", "o", "output"))

	args, err := fp.Parse(ctx.Args)
	if err != nil {
		return ctx.CommandError(err)
	}
	var gz bool
	if args["gzip"].(bool) {
		gz = true
	}
	out := ctx.Channel.ID
	if args["out"].(string) != "" {
		channel, err := ctx.ParseChannel(args["out"].(string))
		if err != nil {
			return ctx.CommandError(err)
		}
		out = channel.ID
	}
	ctx.Session.ChannelTyping(out)

	terms, err := c.db.GetTerms(0)
	if err != nil {
		return ctx.CommandError(err)
	}

	export.Terms = terms

	b, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return ctx.CommandError(err)
	}
	fn := fmt.Sprintf("export-%v.json", time.Now().Format("2006-01-02-15-04-05"))

	var buf *bytes.Buffer
	if gz {
		buf = new(bytes.Buffer)
		zw := gzip.NewWriter(buf)
		zw.Name = fn
		_, err = zw.Write(b)
		if err != nil {
			return ctx.CommandError(err)
		}
		err = zw.Close()
		if err != nil {
			return ctx.CommandError(err)
		}
		fn = fn + ".gz"
	} else {
		buf = bytes.NewBuffer(b)
	}

	file := discordgo.File{
		Name:   fn,
		Reader: buf,
	}

	_, err = ctx.Session.ChannelMessageSendComplex(out, &discordgo.MessageSend{
		Content: fmt.Sprintf("%v\n> Done! Archive of %v terms, invoked by %v at %v.", ctx.Author.Mention(), len(terms), ctx.Author.String(), time.Now().Format(time.RFC3339)),
		Files:   []*discordgo.File{&file},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeUsers,
			},
		},
	})
	return err
}
