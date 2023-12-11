package server

import (
	"github.com/shamank/edutour-backend/apigateway/internal/config"
	"github.com/valyala/fasthttp"
	"strconv"
)

type Server struct {
	addr       string
	httpServer *fasthttp.Server
}

func NewServer(httpConfig config.HTTPServer, handler fasthttp.RequestHandler) *Server {
	return &Server{
		addr: httpConfig.Host + ":" + strconv.Itoa(httpConfig.Port),
		httpServer: &fasthttp.Server{
			Handler:      handler,
			WriteTimeout: httpConfig.WriteTimeout,

			ReadTimeout: httpConfig.ReadTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe(s.addr)
}

func (s *Server) Stop() error {
	return s.httpServer.Shutdown()
}
