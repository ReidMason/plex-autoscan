package config

import (
	"encoding/json"
	"os"

	"github.com/ReidMason/plex-autoscan/internal/logger"
)

type Config struct {
	PlexHost  string `json:"plexHost"`
	PlexToken string `json:"plexToken"`
	PlexPort  int    `json:"plexPort"`
}

func NewConfig(log logger.Logger) (Config, error) {
	return loadConfig(log)
}

func loadConfig(log logger.Logger) (Config, error) {
	config := Config{}

	log.Info("Loading config from file")
	file, err := os.Open("data/config.json")
	defer file.Close()
	if err != nil {
		log.Error("Failed to open config file", err)
		return config, err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Error("Failed to decode config file", err)
		return config, err
	}

	return config, nil
}
