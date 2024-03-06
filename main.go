package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/ReidMason/plex-autoscan/internal/config"
	"github.com/ReidMason/plex-autoscan/internal/plex"
	"github.com/ReidMason/plex-autoscan/internal/sonarr"
	"github.com/labstack/echo"
)

func main() {
	var config = config.LoadConfig()

	client := &http.Client{}
	plexService := plex.NewPlex()
	log.Println("Initializing plex")

	err := plexService.Initialize(config.PlexToken, config.PlexHost, config.PlexPort, client)
	if err != nil {
		log.Fatal("Failed to initialize plex", err)
	}

	e := echo.New()

	e.POST("/notify/:service", func(c echo.Context) error {
		serviceName := c.Param("service")

		log.Println("Request recieved from " + serviceName)
		var body sonarr.SonarrWebhookBody
		err := c.Bind(&body)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid request body")
		}

		log.Println("Event type:", body.EventType)
		if body.EventType == "Test" {
			return c.String(http.StatusOK, "Test request received from "+serviceName)
		}

		sonarrPath := body.Series.Path
		plexPath := replacePath(sonarrPath)

		libraries, err := plexService.GetLibraries()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to get libraries")
		}

		// Find relevant library ids
		libraryIds := make([]string, 0)
		log.Println("Plex path:", plexPath)
		for _, library := range libraries {
			for _, location := range library.Locations {
				if strings.HasPrefix(plexPath, location.Path) {
					libraryIds = append(libraryIds, library.Key)
					break
				}
			}
		}

		// Get season number
		var seasonNumber *int = nil
		if len(body.Episodes) > 0 {
			seasonNumber = &body.Episodes[0].SeasonNumber
		}

		for _, libraryId := range libraryIds {
			err = plexService.RefreshSeason(libraryId, plexPath, seasonNumber)
		}

		return c.String(http.StatusOK, "Request recieved from "+serviceName)
	})

	e.Logger.Fatal(e.Start(":3030"))
}

func replacePath(path string) string {
	return strings.Replace(path, "/tv", "/data", 1)
}
