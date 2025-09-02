package main

import (
	"errors"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/go-ini/ini"
)

const defaultConfigFileName = "aoi-dono/aoi-dono.ini"

type config struct {
	DefaultLanguage string
	Mastodon        `ini:"Mastodon"`
	Bluesky         `ini:"Bluesky"`
}
type Mastodon struct {
	Server       string
	ClientID     string
	ClientSecret string
	AccessToken  string
}

type Bluesky struct {
	ServiceUrl string
	Identifier string
	Password   string
}

var ErrNoConfig = errors.New("resource not found")

func (c *config) LoadConfig() error {
	file, err := xdg.ConfigFile(defaultConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}
	cfg, err := ini.LooseLoad(file)
	if err != nil {
		return fmt.Errorf("loose load failed: %w", err)
	}
	err = cfg.MapTo(c)
	if err != nil {
		return fmt.Errorf("could not map ini to struct: %w", err)
	}
	return nil
}

func (c *config) SaveConfig() error {
	file, err := xdg.ConfigFile(defaultConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}
	cfg := ini.Empty()
	err = cfg.ReflectFrom(c)
	if err != nil {
		return fmt.Errorf("failed to create ini file from struct: %w", err)
	}
	err = cfg.SaveTo(file)
	if err != nil {
		return fmt.Errorf("could not save config file: %w", err)
	}
	return nil
}
