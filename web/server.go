package web

import (
	"github.com/SilverCory/CovidSim"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Server struct {
	l     zerolog.Logger
	query CovidSim.QueryStorage
	e     *gin.Engine
}

func NewServer(l zerolog.Logger, query CovidSim.QueryStorage) (Server, error) {
	var ret = Server{
		l:     l,
		query: query,
	}

	e := gin.Default()
	v1 := e.Group("/api/v1/")
	{
		v1.GET("/tree/:user", ret.v1TreeGETHandle)
	}

	v1User := e.Group("/api/v1/user/:user")
	{
		v1User.GET("/", ret.v1UserGETHandle)
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
