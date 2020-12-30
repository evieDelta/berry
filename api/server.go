package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Starshine113/termbot/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type router struct {
	db    *db.Db
	conf  conf
	sugar *zap.SugaredLogger
}

type conf struct {
	DatabaseURL string `yaml:"database_url"`
	Port        string `yaml:"port"`
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	var c conf

	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(configFile, &c)
	log.Println("Loaded configuration file.")

	d, err := db.Init(c.DatabaseURL, sugar)
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	sugar.Info("Connected to database.")

	r := router{db: d, conf: c, sugar: sugar}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/api/v1/search/:term", r.search)
	e.GET("/api/v1/term/:id", r.term)

	// get port
	port := c.Port
	strings.TrimPrefix(port, ":")
	if port == "" {
		port = "1300"
	}

	go func() {
		if err := e.Start(":" + c.Port); err != nil {
			sugar.Info("Shutting down server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		sugar.Fatal(err)
	}
}
