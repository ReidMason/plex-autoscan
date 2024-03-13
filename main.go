package main

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/ReidMason/plex-autoscan/internal/config"
	"github.com/ReidMason/plex-autoscan/internal/logger"
	"github.com/ReidMason/plex-autoscan/internal/plex"
	"github.com/ReidMason/plex-autoscan/internal/sonarr"
	"github.com/labstack/echo"
)

func main() {
	log := logger.NewLogger()

	log.Info("Starting plex-autoscan")

	config, err := config.NewConfig(log)
	if err != nil {
		panic(err)
	}

	plexService := plex.NewPlex(log)

	client := &http.Client{}
	err = plexService.Initialize(config.PlexToken, config.PlexHost, config.PlexPort, client)
	if err != nil {
		log.Error("Failed to initialize plex service")
		return
	}

	e := echo.New()

	e.POST("/notify/:service", func(c echo.Context) error {
		serviceName := c.Param("service")

		log.Info("Received request", slog.String("ServiceName", serviceName))
		var body sonarr.SonarrWebhookBody
		err := c.Bind(&body)
		if err != nil {
			body := c.Request().Body
			log.Error("Failed to bind request body", slog.Any("error", err), slog.Any("body", body))
			return c.String(http.StatusBadRequest, "Invalid request body")
		}

		log.Debug("Received body", slog.Any("body", body))

		if body.EventType == "Test" {
			log.Info("Test request received from " + serviceName)
			return c.String(http.StatusOK, "Test request received from "+serviceName)
		}

		sonarrPath := body.Series.Path
		plexPath := sonarrPath
		remappings := config.Remappings[serviceName]
		for _, remapping := range remappings {
			plexPath = replacePath(sonarrPath, remapping)
		}

		libraries, err := plexService.GetLibraries()
		if err != nil {
			log.Error("Failed to get libraries", slog.Any("error", err))
			return c.String(http.StatusInternalServerError, "Failed to get libraries")
		}

		// Find relevant library ids
		log.Info("Received path", slog.String("sonarrPath", sonarrPath), slog.String("plexPath", plexPath))
		libraryIds := make([]string, 0)
		for _, library := range libraries {
			for _, location := range library.Locations {
				if strings.HasPrefix(plexPath, location.Path) {
					libraryIds = append(libraryIds, library.Key)
					break
				}
			}
		}

		if len(libraryIds) == 0 {
			log.Error("No libraries found for path", slog.String("path", plexPath))
			return c.String(http.StatusBadRequest, "No library found for path")
		}

		// Get season number
		var seasonNumber *int = nil
		if len(body.Episodes) > 0 {
			log.Info("Found a season number", slog.Int("seasonNumber", body.Episodes[0].SeasonNumber))
			seasonNumber = &body.Episodes[0].SeasonNumber
		}

		for _, library := range libraryIds {
			err = plexService.RescanPath(library, plexPath, seasonNumber)
		}

		return c.String(http.StatusOK, "Request recieved from "+serviceName)
	})

	e.Logger.Fatal(e.Start("localhost:3030"))
}

func replacePath(path string, remapping config.Remapping) string {
	return strings.Replace(path, remapping.From, remapping.To, 1)
}
