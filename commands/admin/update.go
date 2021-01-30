package admin

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/starshine-sys/bcr"
)

func (c *Admin) update(ctx *bcr.Context) (err error) {
	_, err = ctx.Send("Updating Git repository...", nil)
	if err != nil {
		return err
	}

	git := exec.Command("git", "pull")
	pullOutput, err := git.Output()
	if err != nil {
		_, err = ctx.Send(fmt.Sprintf("Error pulling repository:\n```%v```", err), nil)
		return err
	}
	_, err = ctx.Send(fmt.Sprintf("Git:\n```%v```", string(pullOutput)), nil)
	if err != nil {
		return err
	}

	t := time.Now()
	update := exec.Command("/usr/local/go/bin/go", "build", "-v")
	updateOutput, err := update.Output()
	if err != nil {
		_, err = ctx.Send(fmt.Sprintf("Error building:\n```%v```", err), nil)
		return err
	}
	buildTime := time.Since(t).Round(time.Millisecond)
	_, err = ctx.Send(fmt.Sprintf("`go build` (%v):\n```%v```", buildTime, bcr.DefaultValue(string(updateOutput), "[no output]")), nil)
	return
}

func (c *Admin) restart(ctx *bcr.Context) (err error) {
	var silent bool

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVarP(&silent, "silent", "s", false, "If this flag is used, don't set the bot's status")
	err = fs.Parse(ctx.Args)
	if err != nil {
		return c.db.InternalError(ctx, err)
	}
	ctx.Args = fs.Args()

	if len(ctx.Args) > 0 {
		t, err := time.ParseDuration(ctx.Args[0])
		if err == nil {
			c.sugar.Infof("Restart scheduled in %v by %v#%v (%v)", t.Round(time.Second), ctx.Author.Username, ctx.Author.Discriminator, ctx.Author.ID)

			_, err = ctx.Sendf("Restart scheduled in %v.", t.Round(time.Second))
			if err != nil {
				c.sugar.Error("Error sending message:", err)
			}

			// set status
			if !silent {
				c.stopStatus <- true
				ctx.Session.Gateway.UpdateStatus(gateway.UpdateStatusData{
					Status: gateway.OnlineStatus,
					Activities: &[]discord.Activity{{
						Name: fmt.Sprintf("⏲️ Restart scheduled for %v", time.Now().UTC().Add(t).Format("15:04:05 MST")),
					}},
				})
			}

			time.Sleep(t)
		}
	}

	ctx.Session.Gateway.UpdateStatus(gateway.UpdateStatusData{
		Status: gateway.IdleStatus,
		Activities: &[]discord.Activity{{
			Name: "Restarting, please wait...",
		}},
	})

	_, err = ctx.Send("Restarting the bot, please wait...", nil)
	if err != nil {
		return err
	}
	c.sugar.Infof("Restart command received from %v#%v (%v), shutting down...", ctx.Author.Username, ctx.Author.Discriminator, ctx.Author.ID)

	ctx.Router.Session.Close()
	c.sugar.Infof("Disconnected from Discord.")
	c.db.Pool.Close()
	c.sugar.Infof("Closed database connection.")
	os.Exit(0)
	return nil
}
