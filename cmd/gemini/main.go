package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"

	"git.sr.ht/~adnano/go-gemini/certificate"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	// config
	var c conf
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		sugar.Fatal(err)
	}
	err = yaml.Unmarshal(configFile, &c)
	if err != nil {
		sugar.Fatalf("Error loading configuration file: %v", err)
	}
	sugar.Info("Loaded configuration file.")

	certs := certificate.Store{}
	err = certs.Load(c.CertificatePath)
	certs.SetPath(c.CertificatePath)
	if err != nil {
		sugar.Fatalf("Error loading certificates: %v", err)
	}
	if c.BaseURL != "" {
		certs.Register(strings.SplitN(c.BaseURL, ":", 2)[0])
	} else {
		certs.Register("*")
	}

	// startup
	s := &site{sugar: sugar, conf: c, certs: certs.Get}
	s.Run()

	// wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	if s.close != nil {
		s.close()
		s.close = nil
	}

	// shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.gemini.Shutdown(ctx); err != nil {
		s.sugar.Fatal(err)
	}
}
