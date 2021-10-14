package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/gateway/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/utils/wsutil"
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
	dbsearch "github.com/termora/berry/db/search"
	"github.com/termora/berry/db/search/typesense"
)

var debug, disableEventLoop, moreDebug bool

func init() {
	pflag.BoolVarP(&debug, "debug", "d", true, "Debug logging")
	pflag.BoolVarP(&disableEventLoop, "noloop", "N", false, "Disable event loop that will kill bot after 5 minutes of no events")
	pflag.BoolVarP(&moreDebug, "more-debug", "", false, "Even MORE debug logs (very spammy)")
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
	d.Config = &c
	d.TermBaseURL = c.TermBaseURL()
	defer func() {
		d.Pool.Close()
		sugar.Infof("Closed database connection.")
	}()

	if c.Auth.TypesenseURL != "" && c.Auth.TypesenseKey != "" {
		d.Searcher, err = typesense.New(c.Auth.TypesenseURL, c.Auth.TypesenseKey, d.Pool, db.Debug)
		if err != nil {
			sugar.Fatalf("Error connecting to Typesense: %v", err)
		}
	}

	// sync terms
	terms, err := d.GetTerms(dbsearch.FlagSearchHidden)
	if err != nil {
		sugar.Fatalf("Couldn't fetch all terms: %v", err)
	}

	err = d.SyncTerms(terms)
	if err != nil {
		sugar.Fatalf("Couldn't synchronize terms: %v", err)
	}
	sugar.Info("Synchronized terms with search instance!")

	sugar.Info("Connected to database.")

	// create a new state
	b, err := bcrbot.New(c.Auth.Token)
	if err != nil {
		sugar.Fatalf("Error creating bot: %v", err)
	}
	b.Router.ShardManager.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		state.Cabinet.MessageStore = store.Noop
		state.Gateway.ErrorLog = func(err error) {
			sugar.Errorf("Gateway error: %v", err)
		}
	})

	b.Owner(c.Bot.BotOwners...)

	// set the default embed colour and blacklist function
	b.Router.EmbedColor = db.EmbedColour
	b.Router.BlacklistFunc = d.CtxInBlacklist

	// create the bot instance
	bot := bot.New(
		b, sugar, &c, d, hub)
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

	state, _ := bot.Router.StateFromGuildID(0)
	botUser, _ := state.Me()
	bot.Router.Bot = botUser
	bot.Router.Prefixes = append(bot.Router.Prefixes, fmt.Sprintf("<@%v>", bot.Router.Bot.ID), fmt.Sprintf("<@!%v>", bot.Router.Bot.ID))

	// open a connection to Discord
	if err = bot.Start(context.Background()); err != nil {
		sugar.Fatal("Failed to connect:", err)
	}

	// Defer this to make sure that things are always cleanly shutdown even in the event of a crash
	defer func() {
		bot.Router.ShardManager.Close()
		sugar.Infof("Disconnected from Discord.")
	}()

	sugar.Info("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")
	sugar.Infof("User: %v (%v)", botUser.Tag(), botUser.ID)

	if c.Bot.SlashCommands.Enabled {
		if len(c.Bot.SlashCommands.Guilds) > 0 {
			sugar.Infof("Syncing commands in %v...", c.Bot.SlashCommands.Guilds)
		} else {
			sugar.Info("Syncing slash commands...")
		}
		err = bot.Router.SyncCommands(c.Bot.SlashCommands.Guilds...)
		if err != nil {
			sugar.Errorf("Couldn't sync commands: %v", err)
		} else {
			sugar.Info("Synced commands!")
		}
	}

	go timer(sugar)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	exitCh := make(chan struct{})
	if !disableEventLoop {
		eventCh := make(chan gateway.Event, 100)

		go eventThing(sugar, eventCh, exitCh)

		bot.Router.AddHandler(eventCh)
	}

	shutdownFromNoEvents := false
	select {
	case <-ctx.Done():
	case <-exitCh:
		shutdownFromNoEvents = true
	}

	sugar.Infof("Interrupt signal received. Shutting down...")

	if c.Bot.StartStopLog.ID.IsValid() {
		wh := webhook.New(c.Bot.StartStopLog.ID, c.Bot.StartStopLog.Token)

		t := time.Now().UTC()
		s := t.Unix()

		wh.Execute(webhook.ExecuteData{
			Username:  botUser.Username,
			AvatarURL: botUser.AvatarURL(),
			Content:   fmt.Sprintf("Shutting down at <t:%v:D> <t:%v:T>\nShutting down due to no events? %v", s, s, shutdownFromNoEvents),
		})
	}
}

func timer(sugar *zap.SugaredLogger) {
	t := time.Now().UTC()
	ch := time.Tick(10 * time.Minute)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	for {
		select {
		case <-ch:
			sugar.Debugf("Tick received, %s since last tick.", time.Since(t))
			t = time.Now().UTC()
		case <-ctx.Done():
			return
		}
	}
}

func eventThing(s *zap.SugaredLogger, ch <-chan gateway.Event, out chan<- struct{}) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	t := time.AfterFunc(5*time.Minute, func() {
		out <- struct{}{}
	})

	for {
		select {
		case ev := <-ch:
			if moreDebug {
				s.Debugf("Received event %s", reflect.ValueOf(ev).Elem().Type().Name())
			}
			t.Stop()
			t = time.AfterFunc(5*time.Minute, func() {
				out <- struct{}{}
			})
		case <-ctx.Done():
			// break if we're shutting down
			break
		}
	}
}
