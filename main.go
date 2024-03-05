package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	plex "github.com/ReidMason/plex-autoscan/internal"
	"github.com/labstack/echo"
)

type SonarrWebhookBody struct {
	EventType      string `json:"eventType"`
	InstanceName   string `json:"instanceName"`
	ApplicationURL string `json:"applicationUrl"`
	Episodes       []struct {
		Title         string `json:"title"`
		ID            int    `json:"id"`
		EpisodeNumber int    `json:"episodeNumber"`
		SeasonNumber  int    `json:"seasonNumber"`
		SeriesID      int    `json:"seriesId"`
		TvdbID        int    `json:"tvdbId"`
	} `json:"episodes"`
	Series struct {
		Title    string `json:"title"`
		Path     string `json:"path"`
		Type     string `json:"type"`
		ID       int    `json:"id"`
		TvdbID   int    `json:"tvdbId"`
		TvMazeID int    `json:"tvMazeId"`
		Year     int    `json:"year"`
	} `json:"series"`
}

type Config struct {
	PlexHost  string `json:"plexHost"`
	PlexToken string `json:"plexToken"`
	PlexPort  int    `json:"plexPort"`
}

func loadConfig() Config {
	file, err := os.Open("data/config.json")
	if err != nil {
		log.Fatal("Failed to open config file", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Failed to decode config file", err)
	}

	return config
}

func main() {
	var config = loadConfig()

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
		var body SonarrWebhookBody
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
