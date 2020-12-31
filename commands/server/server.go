package server

import (
	"github.com/Starshine113/crouter"
	"github.com/Starshine113/berry/db"
	"github.com/bwmarrin/discordgo"
)

type commands struct {
	db *db.Db
}

// Init ...
func Init(db *db.Db, r *crouter.Router) {
	c := &commands{db: db}

	g := r.AddGroup(&crouter.Group{
		Name:        "blacklist",
		Description: "Manage this server's blacklist",
		Command: &crouter.Command{
			Name:    "Show",
			Summary: "Show the current blacklist",

			Permissions: discordgo.PermissionManageServer,
			Command:     c.blacklist,
		},
	})

	g.AddCommand(&crouter.Command{
		Name:        "add",
		Description: "Add a channel to the blacklist",
		Usage:       "<channel>",

		Permissions: discordgo.PermissionManageServer,
		Command:     c.blacklistAdd,
	})

	g.AddCommand(&crouter.Command{
		Name:        "remove",
		Description: "Remove a channel from the blacklist",
		Usage:       "<channel>",

		Permissions: discordgo.PermissionManageServer,
		Command:     c.blacklistRemove,
	})
}
