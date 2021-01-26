package static

import (
	"fmt"
	"math"
	"os/exec"
	"runtime"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/db"
)

var botVersion = "v0.4"
var gitVer string

func init() {
	git := exec.Command("git", "rev-parse", "--short", "HEAD")
	b, _ := git.Output()
	gitVer = string(b)
	if gitVer == "" {
		gitVer = "unknown"
	}
}

func (c *Commands) about(ctx *bcr.Context) (err error) {
	c.cmdMutex.RLock()
	defer c.cmdMutex.RUnlock()
	embed := &discord.Embed{
		Title: "About",
		Color: db.EmbedColour,
		Footer: &discord.EmbedFooter{
			Text: "Made with Arikawa",
		},
		Thumbnail: &discord.EmbedThumbnail{
			URL: ctx.Bot.AvatarURL(),
		},
		Timestamp: discord.NewTimestamp(time.Now()),
		Fields: []discord.EmbedField{
			{
				Name:   "Bot version",
				Value:  fmt.Sprintf("%v-%v (bcr v%v)", botVersion, gitVer, bcr.Version()),
				Inline: true,
			},
			{
				Name:   "Go version",
				Value:  runtime.Version(),
				Inline: true,
			},
			{
				Name:   "Invite",
				Value:  fmt.Sprintf("[Invite link](%v)", invite(ctx)),
				Inline: true,
			},
			{
				Name: "Uptime",
				Value: fmt.Sprintf(
					"%v\n(Since %v)\n\n**Terms:** %v\n",
					prettyDurationString(time.Since(c.start)),
					c.start.Format("Jan _2 2006, 15:04:05 MST"),
					c.DB.TermCount(),
				),
				Inline: false,
			},
			{
				Name:   "Credits",
				Value:  fmt.Sprintf("Check `%vcredits`!", ctx.Router.Prefixes[0]),
				Inline: true,
			},
			{
				Name:   "Source code",
				Value:  "[GitHub](https://github.com/starshine-sys/berry)\n/ Licensed under the [GNU AGPLv3](https://www.gnu.org/licenses/agpl-3.0.html)",
				Inline: true,
			},
		},
	}

	_, err = ctx.Send("", embed)
	return
}

func invite(ctx *bcr.Context) string {
	// perms is the list of permissions the bot will be granted by default
	var perms = discord.PermissionViewChannel +
		discord.PermissionReadMessageHistory +
		discord.PermissionSendMessages +
		discord.PermissionManageMessages +
		discord.PermissionEmbedLinks +
		discord.PermissionAttachFiles +
		discord.PermissionUseExternalEmojis +
		discord.PermissionAddReactions

	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%v&permissions=%v&scope=applications.commands%%20bot", ctx.Bot.ID, perms)
}

func prettyDurationString(duration time.Duration) (out string) {
	var days, hours, hoursFrac, minutes float64

	hours = duration.Hours()
	hours, hoursFrac = math.Modf(hours)
	minutes = hoursFrac * 60

	hoursFrac = math.Mod(hours, 24)
	days = (hours - hoursFrac) / 24
	hours = hours - (days * 24)
	minutes = minutes - math.Mod(minutes, 1)

	if days != 0 {
		out += fmt.Sprintf("%v days, ", days)
	}
	if hours != 0 {
		out += fmt.Sprintf("%v hours, ", hours)
	}
	out += fmt.Sprintf("%v minutes", minutes)

	return
}
