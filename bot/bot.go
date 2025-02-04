// Package bot contains the bot's core functionality.
package bot

import (
	"context"
	"sort"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
	"github.com/getsentry/sentry-go"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mediocregopher/radix/v4"
	"github.com/starshine-sys/bcr"
	bcrbot "github.com/starshine-sys/bcr/bot"
	"github.com/termora/berry/common"
	"github.com/termora/berry/common/log"
	"github.com/termora/berry/db"
	"github.com/termora/berry/helper"
)

// Bot is the main bot struct
type Bot struct {
	*bcrbot.Bot
	Config common.Config
	DB     *db.DB

	Sentry    *sentry.Hub
	UseSentry bool

	Guilds   map[discord.GuildID]discord.Guild
	GuildsMu sync.Mutex

	Stats *StatsClient

	Helper *helper.Helper

	redis radix.Client
}

// New creates a new instance of Bot
func New(
	bot *bcrbot.Bot,
	config common.Config,
	db *db.DB, hub *sentry.Hub) *Bot {
	b := &Bot{
		Bot:       bot,
		Config:    config,
		DB:        db,
		Sentry:    hub,
		UseSentry: hub != nil,
		Guilds:    map[discord.GuildID]discord.Guild{},
	}

	if config.Core.Redis != "" {
		client, err := (&radix.PoolConfig{}).New(context.Background(), "tcp", config.Core.Redis)
		if err == nil {
			b.redis = client
		}
	}

	// set the router's prefixer
	b.Router.Prefixer = b.Prefixer

	// add the required handlers
	b.Router.ShardManager.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		state.PreHandler = handler.New()
		state.AddHandler(b.MessageCreate)
		state.AddHandler(b.InteractionCreate)
		state.AddHandler(b.GuildCreate)
		state.AddHandler(b.reminderInteraction) // TODO: remove once message content intent launches
		state.PreHandler.AddSyncHandler(b.GuildDelete)
	})

	// setup stats if metrics are enabled
	b.setupStats()

	if config.Bot.SupportToken != "" {
		h, err := helper.New(config.Bot.SupportToken, config.Bot.SupportGuildID, db)
		if err != nil {
			log.Errorf("Error creating helper: %v", err)
		}
		b.Helper = h
	}

	return b
}

// Add adds a module to the bot
func (bot *Bot) Add(f func(*Bot) (string, []*bcr.Command)) {
	m, c := f(bot)

	// sort the list of commands
	sort.Sort(bcr.Commands(c))

	// add the module
	bot.Modules = append(bot.Modules, &botModule{
		name:     m,
		commands: c,
	})
}

// Report reports an exception to Sentry if that's used, and the error is "our problem"
func (bot *Bot) Report(ctx *bcr.Context, err error) *sentry.EventID {
	if db.IsOurProblem(err) && bot.UseSentry {
		return bot.DB.CaptureError(ctx, err)
	}
	return nil
}

func (bot *Bot) guildCount() int {
	bot.GuildsMu.Lock()
	count := len(bot.Guilds)
	bot.GuildsMu.Unlock()
	return count
}

func (bot *Bot) setupStats() {
	if bot.Config.Bot.InfluxDB.URL != "" {
		log.Infof("Setting up InfluxDB client")

		bot.Stats = &StatsClient{
			Client:     influxdb2.NewClient(bot.Config.Bot.InfluxDB.URL, bot.Config.Bot.InfluxDB.Token).WriteAPI(bot.Config.Bot.InfluxDB.Org, bot.Config.Bot.InfluxDB.Bucket),
			guildCount: bot.guildCount,
		}

		bot.Router.ShardManager.ForEach(func(s shard.Shard) {
			state := s.(*state.State)

			state.Client.Client.OnResponse = append(state.Client.Client.OnResponse, func(httpdriver.Request, httpdriver.Response) error {
				go bot.Stats.IncAPICall()
				return nil
			})
		})

		bot.DB.IncFunc = bot.Stats.IncQuery

		go bot.Stats.submit()
	}
}
