package search

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/Starshine113/crouter"
)

func (c *commands) term(ctx *crouter.Ctx) (err error) {
	if err = ctx.CheckMinArgs(1); err != nil {
		_, err = ctx.Sendf("❌ No term ID provided.")
		return
	}
	channel := ctx.Channel
	if len(ctx.Args) > 1 {
		channel, err = ctx.ParseChannel(strings.Join(ctx.Args[1:], " "))
		if err != nil {
			c.Sugar.Error("Error getting channel:", err)
		}
	}
	if channel.GuildID != ctx.Message.GuildID {
		_, err = ctx.Sendf("❌ The channel you gave is not in this server.")
		return
	}

	id, err := strconv.Atoi(ctx.Args[0])
	if err != nil {
		_, err = ctx.Sendf("❌ No or invalid ID provided.")
		return
	}

	term, err := c.Db.GetTerm(id)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			_, err = ctx.Sendf("❌ No term with that ID found.")
			return
		}
		c.Sugar.Errorf("Error getting term %v: %v", id, err)
		_, err = ctx.Sendf("❌ Internal error occurred while trying to fetch the requested term.\nIf this issue persists, please contact the bot developer.")
		return
	}

	perms, err := getPerms(ctx, ctx.Author.ID, channel.ID)
	if err != nil {
		c.Sugar.Errorf("Error getting perms for %v in %v: %v", ctx.Author.ID, channel.ID, err)
		_, err = ctx.Sendf("❌ An error occurred while trying to get permissions.\nIf this issue persists, please contact the bot developer.")
		return
	}

	if perms&discordgo.PermissionManageMessages != discordgo.PermissionManageMessages {
		_, err = ctx.Sendf("❌ Error: this command requires the `Manage Messages` permission in the channel you're posting to.")
		return
	}

	_, err = ctx.Session.ChannelMessageSendEmbed(channel.ID, term.TermEmbed(c.conf.Bot.TermBaseURL))
	if err != nil {
		return
	}

	if channel.ID != ctx.Channel.ID {
		_, err = ctx.Sendf("✅ Message sent to %v!", channel.Mention())
	}
	return
}

func getPerms(ctx *crouter.Ctx, userID, channelID string) (perms int, err error) {
	perms, err = ctx.Session.State.UserChannelPermissions(userID, channelID)
	if err == discordgo.ErrStateNotFound {
		perms, err = ctx.Session.UserChannelPermissions(userID, channelID)
		return
	}
	return
}
