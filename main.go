package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/diamondburned/arikawa/v2/state"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/berry/bot"
	"github.com/starshine-sys/berry/commands/admin"
	"github.com/starshine-sys/berry/commands/search"
	"github.com/starshine-sys/berry/commands/server"
	"github.com/starshine-sys/berry/commands/static"
	"github.com/starshine-sys/berry/db"
)

var sugar *zap.SugaredLogger

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.RedirectStdLog(logger)
	sugar = logger.Sugar()

	c := getConfig(sugar)

	d, err := db.Init(c.Auth.DatabaseURL, sugar)
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	d.Config = c
	sugar.Info("Connected to database.")

	s, err := state.NewWithIntents("Bot "+c.Auth.Token, bcr.RequiredIntents)
	if err != nil {
		log.Fatalln("Error creating state:", err)
	}

	r := bcr.NewRouter(s, c.Bot.BotOwners, c.Bot.Prefixes)
	r.EmbedColor = db.EmbedColour

	// set blacklist function
	r.BlacklistFunc = d.CtxInBlacklist

	// create the bot instance
	bot := bot.New(sugar, c, r, d)
	// add search commands
	bot.Add(search.Init)
	// add static commands
	bot.Add(static.Init)
	// add server commands
	bot.Add(server.Init)
	// add admin commands
	bot.Add(admin.Init)

	// open a connection to Discord
	if err = s.Open(); err != nil {
		sugar.Fatal("Failed to connect:", err)
	}

	// Defer this to make sure that things are always cleanly shutdown even in the event of a crash
	defer func() {
		s.Close()
		sugar.Infof("Disconnected from Discord.")
		d.Pool.Close()
		sugar.Infof("Closed database connection.")
	}()

	sugar.Info("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")

	botUser, _ := s.Me()
	sugar.Infof("User: %v#%v (%v)", botUser.Username, botUser.Discriminator, botUser.ID)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	sugar.Infof("Interrupt signal received. Shutting down...")
}
