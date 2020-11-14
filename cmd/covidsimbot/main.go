package main

import (
	"fmt"
	"github.com/SilverCory/CovidSim/cache"
	"github.com/SilverCory/CovidSim/discord"
	"github.com/SilverCory/CovidSim/storage"
	"github.com/caarlos0/env/v6"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Spreading 'rona...")

	var cfg = Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Unable to parse env: %+v\n", err)
		os.Exit(1)
		return
	}

	store, err := storage.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		fmt.Printf("Unable to open storage: %v\n", err)
		os.Exit(1)
		return
	}

	ca, err := cache.NewRedis(cfg.RedisAddress, cfg.RedisUsername, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		fmt.Printf("Unable to open cache: %v\n", err)
		os.Exit(1)
		return
	}

	bot, err := discord.NewBot(
		cfg.DiscordBotToken,
		store,
		ca,
		cfg.DiscordWebhookID,
		cfg.DiscordWebhookToken,
	)
	if err != nil {
		fmt.Printf("Unable to open discord: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Waiting for interrupt.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err := ca.Close(); err != nil {
		fmt.Printf("Error closing cache %v\n", err)
	}

	if err := bot.Close(); err != nil {
		fmt.Printf("Error closing store %v\n", err)
	}
	fmt.Println("Cya!!!")
}
