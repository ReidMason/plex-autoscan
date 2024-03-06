package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	PlexHost  string `json:"plexHost"`
	PlexToken string `json:"plexToken"`
	PlexPort  int    `json:"plexPort"`
}

func LoadConfig() Config {
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
