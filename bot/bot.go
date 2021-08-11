// Package bot contains the bot's core functionality.
package bot

import (
	"sort"

	"github.com/diamondburned/arikawa/v3/gateway/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/getsentry/sentry-go"
	"github.com/starshine-sys/bcr"
	bcrbot "github.com/starshine-sys/bcr/bot"
	"github.com/termora/berry/db"
	"github.com/termora/berry/structs"
	"go.uber.org/zap"
)

// Bot is the main bot struct
type Bot struct {
	*bcrbot.Bot

	Sugar  *zap.SugaredLogger
	Config *structs.BotConfig
	DB     *db.Db

	Sentry    *sentry.Hub
	UseSentry bool

	GuildCount int64
}

// Module is a single module/category of commands
type Module interface {
	String() string
	Commands() []*bcr.Command
}

// New creates a new instance of Bot
func New(
	bot *bcrbot.Bot,
	s *zap.SugaredLogger,
	config *structs.BotConfig,
	db *db.Db, hub *sentry.Hub) *Bot {
	b := &Bot{
		Bot:       bot,
		Sugar:     s,
		Config:    config,
		DB:        db,
		Sentry:    hub,
		UseSentry: hub != nil,
	}

	// set the router's prefixer
	b.Router.Prefixer = b.Prefixer

	// add the required handlers
	b.Router.ShardManager.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		state.PreHandler = handler.New()
		state.PreHandler.Synchronous = true
		state.AddHandler(b.MessageCreate)
		state.AddHandler(b.InteractionCreate)
		state.AddHandler(b.GuildCreate)
		state.PreHandler.AddHandler(b.GuildDelete)
	})

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
