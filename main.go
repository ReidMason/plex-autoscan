package main

import (
	"net/http"

	"github.com/ReidMason/plex-autoscan/api"
	"github.com/ReidMason/plex-autoscan/internal/config"
	"github.com/ReidMason/plex-autoscan/internal/logger"
	"github.com/ReidMason/plex-autoscan/internal/notificationHandler"
	"github.com/ReidMason/plex-autoscan/internal/plex"
)

func main() {
	log := logger.NewLogger()

	log.Info("Starting plex-autoscan")

	config, err := config.NewConfig(log)
	if err != nil {
		panic(err)
	}

	plexService := plex.NewPlex(log)
	notificationHandler := notificationHandler.NewNotificationHandler(plexService, config.Remappings, log)

	client := &http.Client{}
	err = plexService.Initialize(config.PlexToken, config.PlexHost, config.PlexPort, client)
	if err != nil {
		log.Error("Failed to initialize plex service")
		return
	}

	api.NewServer(config, notificationHandler, log).Start()
}
