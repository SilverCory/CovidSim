package web

import (
	"context"
	"fmt"
	"github.com/SilverCory/CovidSim"
	"github.com/SilverCory/CovidSim/config"
	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/rbcervilla/redisstore/v8"
	"github.com/rs/zerolog"
)

type Server struct {
	l          zerolog.Logger
	query      CovidSim.QueryStorage
	redisStore *redisstore.RedisStore
	e          *gin.Engine
}

func NewServer(l zerolog.Logger, query CovidSim.QueryStorage, rdb *redis.Client, cfg config.Config) (Server, error) {
	var ret = Server{
		l:     l,
		query: query,
	}

	redisStore, err := redisstore.NewRedisStore(context.Background(), rdb)
	if err != nil {
		return Server{}, fmt.Errorf("web NewServer: unable to create redis store: %w", err)
	}
	ret.redisStore = redisStore
	ret.redisStore.KeyPrefix("session:")
	ret.redisStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 60,
		HttpOnly: true,
	})

	goth.UseProviders(
		discord.New(
			cfg.DiscordClientID,
			cfg.DiscordSecret,
			"http://127.0.0.1:8080/auth/discord/callback/",
			discord.ScopeIdentify,
			discord.ScopeEmail,
		),
	)

	gothic.Store = ret.redisStore

	e := gin.New()
	e.Use(ginzerolog.Logger("gin"), gin.Recovery())

	v1 := e.Group("/api/v1/")
	{
		v1.GET("/tree/:user", ret.v1TreeGETHandle)
		v1User := v1.Group("/user/:user")
		{
			v1User.GET("/", ret.v1UserGETHandle)
		}
	}

	discordAuth := e.Group("/auth/discord")
	{
		discordAuth.Use(discordOAuthContextSetter)
		discordAuth.GET("/", ret.oauthDiscordIndex)
		discordAuth.GET("/callback", ret.oauthDiscordCallback)
	}

	ret.e = e
	return ret, nil
}

func (s *Server) Start(listenAddr string) error {
	return s.e.Run(listenAddr)
}

func (s *Server) Close() error {
	return nil // TODO
}
