package main

type Config struct {
	MySQLDSN string `env:"MYSQL_DSN"`

	RedisAddress  string `env:"REDIS_ADDR"`
	RedisUsername string `env:"REDIS_USER"`
	RedisPassword string `env:"REDIS_PASS"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	DiscordBotToken     string `env:"BOT_TOKEN"`
	DiscordWebhookID    string `env:"INFECTION_HOOK_ID"`
	DiscordWebhookToken string `env:"INFECTION_HOOK_TOKEN"`
}
