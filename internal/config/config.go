package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Tracker       string `toml:"tracker"`
	TicketUrlStem string `toml:"ticket_url_stem"`
}

func LoadConfig() (*Config, error) {
	filename := ".ghpr.toml"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return &Config{
			Tracker: "none",
		}, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(filename, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}
