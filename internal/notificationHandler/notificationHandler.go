package notificationHandler

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/ReidMason/plex-autoscan/internal/config"
	"github.com/ReidMason/plex-autoscan/internal/logger"
	"github.com/ReidMason/plex-autoscan/internal/plex"
	"github.com/ReidMason/plex-autoscan/internal/sonarr"
)

type NotificationHandler struct {
	log         logger.Logger
	plexService *plex.Plex
	remappings  config.Remappings
}

func NewNotificationHandler(plexService *plex.Plex, remappings config.Remappings, log logger.Logger) *NotificationHandler {
	return &NotificationHandler{log: log, plexService: plexService, remappings: remappings}
}

func (nh NotificationHandler) ProcessNotification(body sonarr.SonarrWebhookBody, serviceName string) error {
	nh.log.Debug("Received body", slog.Any("body", body))

	if body.EventType == "Test" {
		nh.log.Info("Test request received from " + serviceName)
		return nil
	}

	sonarrPath := body.Series.Path
	plexPath := remapPath(sonarrPath, nh.remappings[serviceName])

	libraries, err := nh.plexService.GetLibraries()
	if err != nil {
		nh.log.Error("Failed to get libraries", slog.Any("error", err))
		return errors.New("Failed to get libraries")
	}

	// Find relevant library ids
	nh.log.Info("Received path", slog.String("sonarrPath", sonarrPath), slog.String("plexPath", plexPath))
	libraryIds := findLibraryIds(libraries, plexPath)

	if len(libraryIds) == 0 {
		nh.log.Error("No libraries found for path", slog.String("path", plexPath))
		return errors.New("No library found for path")
	}

	// Get season number
	var seasonNumber *int = nil
	if len(body.Episodes) > 0 {
		nh.log.Info("Found a season number", slog.Int("seasonNumber", body.Episodes[0].SeasonNumber))
		seasonNumber = &body.Episodes[0].SeasonNumber
	}

	for _, library := range libraryIds {
		err = nh.plexService.RescanPath(library, plexPath, seasonNumber)
	}

	return nil
}

func findLibraryIds(libraries []plex.Library, plexPath string) []string {
	libraryIds := make([]string, 0)
	for _, library := range libraries {
		for _, location := range library.Locations {
			if strings.HasPrefix(plexPath, location.Path) {
				libraryIds = append(libraryIds, library.Key)
				break
			}
		}
	}

	return libraryIds
}

func remapPath(path string, remappings []config.Remapping) string {
	plexPath := path
	for _, remapping := range remappings {
		plexPath = strings.Replace(path, remapping.From, remapping.To, 1)
	}

	return plexPath
}
