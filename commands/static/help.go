package static

import (
	"fmt"
	"strings"

	"github.com/Starshine113/bcr"
	"github.com/Starshine113/berry/db"
	"github.com/diamondburned/arikawa/v2/discord"
)

func (c *Commands) help(ctx *bcr.Context) (err error) {
	if ctx.RawArgs == "permissions" || ctx.RawArgs == "perms" {
		_, err = ctx.Send("", &discord.Embed{
			Title: "Required permissions",
			Description: fmt.Sprintf(`%v requires the following permissions to function correctly:
		- **Read Messages** & **Send Messages**: to respond to commands
		- **Read Message History**: for the %vsearch command to work
		- **Manage Messages**: to delete reactions on menus
		- **Embed Links**: to send responses for most commands
		- **Attach Files**: to send a glossary export
		- **Add Reactions**: for menus to work
		- **Use External Emojis**: to use custom emotes in a couple of commands`, ctx.Bot.Username, ctx.Router.Prefixes[0]),
			Color: db.EmbedColour,
		})
		return
	}

	if ctx.RawArgs == "privacy" || ctx.RawArgs == "privacy-policy" {
		_, err = ctx.Send("", &discord.Embed{
			Title: "Privacy",
			Description: fmt.Sprintf(`We're not lawyers, we don't want to write a document that no one will (or even can) read.

		This is the data %v collects:
		
		- Data about commands run: command, arguments, user ID, and channel ID, used exclusively for debugging purposes and automatically removed when the bot's logs are rotated
		- A list of blacklisted channels per server
		
		This is the data %v stores in memory, and which is wiped on a restart:
		
		- Message metadata *for its own messages*
		- Message metadata for messages that trigger commands
		
		This is the data %v does *not* collect:
		
		- Any message contents
		- Any user information
		- Information about messages that do not trigger commands
		
		Additionally, there are daily database backups, which only include a list of blacklisted channels (as well as all terms/explanations).
		
		%v is open source, and its source code is available [on GitHub](https://github.com/Starshine113/Berry). While we cannot *prove* that this is the code powering the bot, we promise that it is.`, ctx.Bot.Username, ctx.Bot.Username, ctx.Bot.Username, ctx.Bot.Username),
			Color: db.EmbedColour,
		})
		return err
	}

	if ctx.RawArgs == "autopost" {
		_, err = ctx.Send("", &discord.Embed{
			Title:       "Autopost",
			Description: "To automatically post terms at a set interval, you can use the following custom command for [YAGPDB.xyz](https://yagpdb.xyz/):\n```{{/* Recommended trigger: Minute/Hourly interval */}}\n\nt!random\n{{deleteResponse 1}}```\nOther bots may have similar functionality; if you need a bot whitelisted for commands, feel free to ask on the support server.",
			Color:       db.EmbedColour,
		})
		return
	}

	e := &discord.Embed{
		Color: db.EmbedColour,
		Title: "Help",
		Fields: []discord.EmbedField{
			{
				Name:  "Prefixes",
				Value: fmt.Sprintf("%v uses the prefixes %v, and %v.", ctx.Bot.Username, strings.Join(ctx.Router.Prefixes[:len(ctx.Router.Prefixes)-2], ", "), ctx.Bot.Mention()),
			},
			{
				Name:  "Bot info",
				Value: "`help`: show a list of commands, and some info about the bot\n`help privacy`: show the bot's privacy policy\n`about`: show more in-depth info about the bot.\n`credits`: a list of people who helped create the bot\n`ping`: check the bot's latency\n`hello`: say hi to the bot!\n`invite`: get an invite link for the bot",
			},
			{
				Name:  "Terms",
				Value: "`search`: search the database for a term (alias: `s`)\n`random`: show a random term (alias: `r`)\n`post`: post a single term by ID. You need the **Manage Messages** permission to use this.",
			},
			{
				Name:  "Explanations",
				Value: "`explain`: get a list of all registered explanations (aliases: `e`, `ex`)\n`explain <topic>`: explain the given topic",
			},
			{
				Name:  "Autoposting",
				Value: fmt.Sprintf("%v can't automatically post terms yet, sorry! However, a couple of bots are whitelisted and can trigger commands, which can be used to emulate an autopost function. See `help autopost` for more info.", ctx.Bot.Username),
			},
			{
				Name:  "For staff",
				Value: fmt.Sprintf("You can blacklist most commands, with the exception of `explain`, using the following commands:\n`blacklist`: show the current blacklist\n`blacklist add`: add a channel to the blacklist\n`blacklist remove`: remove a channel from the blacklist\n\nTo stop %v from responding in a channel completely, deny it the \"Read Messages\" permission in that channel.", ctx.Bot.Username),
			},
		},
	}
	if c.config.Bot.ServerInvite != "" {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Support server",
			Value: fmt.Sprintf("Use this link to join the support server, for bot questions and term additions/requests: %v", c.config.Bot.ServerInvite),
		})
	}

	// if custom help fields are defined, add those
	if len(c.config.Bot.HelpFields) != 0 {
		for _, f := range c.config.Bot.HelpFields {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:  f.Name,
				Value: f.Value,
			})
		}
	}

	_, err = ctx.Send("", e)
	return err
}

func (c *Commands) cmdInvite(ctx *bcr.Context) (err error) {
	_, err = ctx.Sendf("Use this link to invite me to your server: <%v>\n\nYou can use the `%vhelp permissions` command to get a detailed explanation of all permissions required.", invite(ctx), ctx.Router.Prefixes[0])
	return
}
