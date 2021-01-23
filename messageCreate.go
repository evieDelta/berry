package main

import (
	"fmt"

	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/structs"
	"github.com/diamondburned/arikawa/v2/gateway"
	"go.uber.org/zap"
)

type messageCreate struct {
	r     *bcr.Router
	c     *structs.BotConfig
	sugar *zap.SugaredLogger
}

func (mc *messageCreate) messageCreate(m *gateway.MessageCreateEvent) {
	var err error

	// defer panic handling
	defer func() {
		r := recover()
		if r != nil {
			mc.sugar.Errorf("Caught panic in channel ID %v (user %v, guild %v): %v", m.ChannelID, m.Author.ID, m.GuildID, err)
		}
	}()

	if mc.r.Bot == nil {
		err = mc.r.SetBotUser()
		if err != nil {
			mc.sugar.Error("Error setting bot user:", err)
			return
		}
		mc.r.Prefixes = append(mc.r.Prefixes, fmt.Sprintf("<@%v>", mc.r.Bot.ID), fmt.Sprintf("<@!%v>", mc.r.Bot.ID))
	}

	// if message was sent by a bot return, unless it's in the list of allowed bots
	if m.Author.Bot && !inSlice(mc.c.Bot.AllowedBots, m.Author.ID.String()) {
		return
	}

	// get context
	ctx, err := mc.r.NewContext(m.Message)
	if err != nil {
		sugar.Error("Error creating context:", err)
		return
	}

	// check if the message might be a command
	if mc.r.MatchPrefix(m.Content) {
		mc.r.Execute(ctx)
	}
}

func inSlice(slice []string, s string) bool {
	for _, i := range slice {
		if i == s {
			return true
		}
	}
	return false
}
