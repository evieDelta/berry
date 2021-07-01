package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v2/state/store"
	"github.com/diamondburned/arikawa/v2/utils/wsutil"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/spf13/pflag"
	bcrbot "github.com/starshine-sys/bcr/bot"
	"github.com/termora/berry/bot"
	"github.com/termora/berry/commands/admin"
	"github.com/termora/berry/commands/pronouns"
	"github.com/termora/berry/commands/search"
	"github.com/termora/berry/commands/server"
	"github.com/termora/berry/commands/static"
	"github.com/termora/berry/db"
)

var (
	shard int
	debug bool
)

func init() {
	pflag.IntVarP(&shard, "shard", "s", 0, "Shard number")
	pflag.BoolVarP(&debug, "debug", "d", true, "Debug logging")
	pflag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// set up a logger
	zcfg := zap.NewProductionConfig()
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zcfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zcfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	if debug {
		zcfg.Level.SetLevel(zapcore.DebugLevel)
	} else {
		zcfg.Level.SetLevel(zapcore.InfoLevel)
	}

	logger, err := zcfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}

	zap.RedirectStdLog(logger)
	sugar := logger.Sugar()

	c := getConfig(sugar)

	if debug {
		wsutil.WSDebug = sugar.Debug
		db.Debug = sugar.Debugf
	}

	// command-line flags, mostly sharding
	c.Shard = shard
	c.Sharded = c.NumShards > 1

	// create a Sentry config
	if c.UseSentry {
		err = sentry.Init(sentry.ClientOptions{
			Dsn: c.Auth.SentryURL,
		})
		if err != nil {
			sugar.Fatalf("sentry.Init: %s", err)
		}
		sugar.Infof("Initialised Sentry")
		// defer this to flush buffered events
		defer sentry.Flush(2 * time.Second)
	}
	hub := sentry.CurrentHub()
	if !c.UseSentry {
		hub = nil
	}

	// connect to the database
	d, err := db.Init(c.Auth.DatabaseURL, sugar)
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	d.SetSentry(hub)
	d.Config = c
	sugar.Info("Connected to database.")

	// create a new state
	b, err := bcrbot.New(c.Auth.Token)
	if err != nil {
		sugar.Fatalf("Error creating bot: %v", err)
	}
	b.Router.State.Cabinet.MessageStore = store.Noop

	b.Router.State.Gateway.ErrorLog = func(err error) {
		sugar.Errorf("Gateway error: %v", err)
	}

	b.Owner(c.Bot.BotOwners...)

	// if the bot is sharded, set the number and count
	if c.Sharded {
		b.Router.State.Gateway.Identifier.SetShard(c.Shard, c.NumShards)
	}

	// set the default embed colour and blacklist function
	b.Router.EmbedColor = db.EmbedColour
	b.Router.BlacklistFunc = d.CtxInBlacklist

	// create the bot instance
	bot := bot.New(
		b, sugar, c, d, hub)
	// add search commands
	bot.Add(search.Init)
	// add pronoun commands
	bot.Add(pronouns.Init)
	// add static commands
	bot.Add(static.Init)
	// add server commands
	bot.Add(server.Init)
	// add admin commands
	bot.Add(admin.Init)

	// open a connection to Discord
	if err = bot.Router.State.Open(); err != nil {
		sugar.Fatal("Failed to connect:", err)
	}

	// Defer this to make sure that things are always cleanly shutdown even in the event of a crash
	defer func() {
		bot.Router.State.Close()
		sugar.Infof("Disconnected from Discord.")
		d.Pool.Close()
		sugar.Infof("Closed database connection.")
	}()

	sugar.Info("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")

	botUser, _ := bot.Router.State.Me()
	sugar.Infof("User: %v#%v (%v)", botUser.Username, botUser.Discriminator, botUser.ID)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	select {
	case <-ctx.Done():
	}

	sugar.Infof("Interrupt signal received. Shutting down...")
}
