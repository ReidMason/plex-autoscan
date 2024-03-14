package api

import (
	"github.com/ReidMason/plex-autoscan/internal/config"
	"github.com/ReidMason/plex-autoscan/internal/logger"
	"github.com/ReidMason/plex-autoscan/internal/notificationHandler"
	notificationhandler "github.com/ReidMason/plex-autoscan/internal/notificationHandler"
	"github.com/labstack/echo"
)

type Server struct {
	log                 logger.Logger
	notificationHandler *notificationHandler.NotificationHandler
	cfg                 config.Config
}

func NewServer(cfg config.Config, notificationHandler *notificationhandler.NotificationHandler, log logger.Logger) *Server {
	return &Server{cfg: cfg, log: log, notificationHandler: notificationHandler}
}

func (s Server) Start() error {
	e := echo.New()

	e.POST("/notify/:service", s.handleNotify)

	e.Logger.Fatal(e.Start("localhost:3030"))

	return nil
}
