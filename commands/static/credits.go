package static

import (
	"fmt"
	"sort"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/termora/berry/db"
)

func (c *Commands) credits(ctx *bcr.Context) (err error) {
	c.memberMu.RLock()
	defer c.memberMu.RUnlock()

	// return if there's no credit fields
	if len(c.Config.Bot.CreditFields) == 0 &&
		(len(c.Config.ContributorRoles) == 0 ||
			len(c.SupportServerMembers) == 0) {
		return nil
	}

	embeds := []discord.Embed{{
		Color:       db.EmbedColour,
		Title:       "Credits",
		Description: fmt.Sprintf("These are the people who helped create %v!", ctx.Bot.Username),
		Fields:      c.Config.Bot.CreditFields,
	}}

	e := discord.Embed{
		Color:       db.EmbedColour,
		Title:       "Contributors",
		Description: fmt.Sprintf("These are the people who have contributed to %v in some capacity!", ctx.Bot.Username),
	}

	for _, role := range c.Config.ContributorRoles {
		members := c.filterByRole(role.ID)
		if len(members) == 0 {
			continue
		}

		var (
			slice []string
			s     string
		)
		for _, m := range members {
			name := m.Nick
			if name == "" {
				name = m.User.Username
			}
			slice = append(slice, name)
		}
		for i, m := range slice {
			if len(s) > 900 {
				s += fmt.Sprintf("\n...and %v others!", len(slice)-i)
				break
			}
			if i != 0 {
				s += ", "
			}
			s += m
		}

		name := role.Name
		if len(slice) != 1 {
			name += "s"
		}
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   name,
			Value:  s,
			Inline: false,
		})
	}
	if len(e.Fields) > 0 {
		embeds = append(embeds, e)
		embeds[0].Description += "\nReact with ➡️ to show everyone who has contributed!"
	}

	_, err = ctx.PagedEmbed(embeds, false)
	return err
}

func (c *Commands) filterByRole(rID discord.RoleID) (members []discord.Member) {
	for _, m := range c.SupportServerMembers {
		for _, id := range m.RoleIDs {
			if id == rID {
				members = append(members, m)
				break
			}
		}
	}

	sort.Slice(members, func(i, j int) bool {
		name1 := members[i].Nick
		name2 := members[j].Nick
		if name1 == "" {
			name1 = members[i].User.Username
		}
		if name2 == "" {
			name2 = members[j].User.Username
		}
		return name1 < name2
	})

	return members
}
