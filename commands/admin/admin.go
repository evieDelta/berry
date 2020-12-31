package admin

import (
	"time"

	"github.com/Starshine113/berry/db"
	"github.com/Starshine113/berry/structs"
	"github.com/Starshine113/crouter"
	"github.com/bwmarrin/discordgo"
)

type commands struct {
	db     *db.Db
	config *structs.BotConfig
}

// Init ...
func Init(db *db.Db, conf *structs.BotConfig, r *crouter.Router) {
	c := &commands{db: db, config: conf}

	r.AddCommand(&crouter.Command{
		Name:        "AddTerm",
		Description: "Add a term",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){c.checkOwner},

		Command: c.addTerm,
	})

	r.AddCommand(&crouter.Command{
		Name:        "DelTerm",
		Description: "Delete a term",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){c.checkOwner},

		Command: c.delTerm,
	})

	r.AddCommand(&crouter.Command{
		Name:        "AddCategory",
		Description: "Add a category",
		Usage:       "<name>",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){c.checkOwner},

		Command: c.addCategory,
	})

	r.AddCommand(&crouter.Command{
		Name:        "AddExplanation",
		Description: "Add an explanation",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){c.checkOwner},

		Command: c.addExplanation,
	})

	r.AddCommand(&crouter.Command{
		Name:        "SetFlags",
		Description: "Set a term's flags",

		CustomPermissions: []func(*crouter.Ctx) (string, bool){c.checkOwner},

		Command: c.setFlags,
	})

	r.AddCommand(&crouter.Command{
		Name:        "Export",
		Description: "Export all terms",
		Usage:       "[-gz] [-channel <ChannelID/Mention>]",

		Cooldown:    time.Minute,
		Permissions: discordgo.PermissionManageMessages,

		Command: c.export,
	})

	r.AddCommand(&crouter.Command{
		Name:        "Guilds",
		Description: "List all guilds the bot is in",

		OwnerOnly: true,
		Command:   c.guilds,
	})
}

func (c *commands) checkOwner(ctx *crouter.Ctx) (string, bool) {
	if c.config.Bot.AdminServer != "" {
		if ctx.Message.GuildID != c.config.Bot.AdminServer {
			return "Bot Admin", false
		}
	}
	for _, id := range c.config.Bot.BotOwners {
		if id == ctx.Author.ID {
			return "", true
		}
	}
	return "Bot Admin", false
}
