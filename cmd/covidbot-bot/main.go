package main

import (
	"fmt"
	"github.com/SilverCory/CovidSim/cache"
	"github.com/SilverCory/CovidSim/config"
	"github.com/SilverCory/CovidSim/discord"
	"github.com/SilverCory/CovidSim/storage"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var l = zerolog.New(os.Stderr).With().Timestamp().Logger()
	l.Info().Msg("Spreading 'rona...")

	cfg, err := config.Get()
	if err != nil {
		l.Error().Err(err).Msg("unable to load configuration.")
		os.Exit(1)
		return
	}

	store, err := storage.NewMySQL(l, cfg.MySQLDSN)
	if err != nil {
		l.Error().Err(err).Msg("unable to open storage.")
		os.Exit(1)
		return
	}

	ca, err := cache.NewRedis(cfg.RedisAddress, cfg.RedisUsername, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		l.Error().Err(err).Msg("unable to open cache.")
		os.Exit(1)
		return
	}

	bot, err := discord.NewBot(
		l,
		cfg.DiscordBotToken,
		store,
		ca,
		cfg.DiscordWebhookID,
		cfg.DiscordWebhookToken,
	)
	if err != nil {
		l.Error().Err(err).Msg("unable to open discord bot.")
		os.Exit(1)
		return
	}

	l.Info().Msg("Done!")
	fmt.Println("Waiting for interrupt.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err := ca.Close(); err != nil {
		l.Error().Err(err).Msg("unable to close cache.")
	}

	if err := bot.Close(); err != nil {
		l.Error().Err(err).Msg("unable to close discord bot.")
	}
	fmt.Println("Cya!!!")
}
