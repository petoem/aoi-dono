package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

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

// parseFlagsAndEnv sets the flags and env values on the config.
//
// 1. Commandline flags, 2. environment variables and 3. config file.
func (c *config) parseFlagsAndEnv() bool {
	// Credentials
	// - Mastodon
	flag.StringVar(&c.Mastodon.Server, "mastodonInstanceUrl", osEnvOrConfigValue("MASTODON_INSTANCE_URL", c.Mastodon.Server), "Mastodon instance URL (e.g., https://mastodon.example)")
	flag.StringVar(&c.Mastodon.AccessToken, "mastodonAccessToken", osEnvOrConfigValue("MASTODON_ACCESS_TOKEN", c.Mastodon.AccessToken), "Mastodon access token")
	flag.StringVar(&c.Mastodon.ClientID, "mastodonClientKey", osEnvOrConfigValue("MASTODON_CLIENT_KEY", c.Mastodon.ClientID), "Mastodon client key")
	flag.StringVar(&c.Mastodon.ClientSecret, "mastodonClientSecret", osEnvOrConfigValue("MASTODON_CLIENT_SECRET", c.Mastodon.ClientSecret), "Mastodon client secret")
	// - Bluesky
	flag.StringVar(&c.Bluesky.ServiceUrl, "blueskyServiceUrl", osEnvOrConfigValue("BLUESKY_SERVICE_URL", c.Bluesky.ServiceUrl), "Bluesky service URL (e.g., https://bsky.social)")
	flag.StringVar(&c.Bluesky.Identifier, "blueskyIdentifier", osEnvOrConfigValue("BLUESKY_IDENTIFIER", c.Bluesky.Identifier), "Bluesky identifier (e.g., @user.bsky.social)")
	flag.StringVar(&c.Bluesky.Password, "blueskyPassword", osEnvOrConfigValue("BLUESKY_PASSWORD", c.Bluesky.Password), "Bluesky password")

	// Other flags
	df := "en"
	if c.DefaultLanguage != "" {
		df = c.DefaultLanguage
	}
	flag.StringVar(&c.DefaultLanguage, "lang", df, "Post language (e.g., jp)")

	shouldSave := flag.Bool("saveToConfig", false, "Save current config to file")
	flag.Parse()
	return *shouldSave
}

func osEnvOrConfigValue(env, defaultValue string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	}
	return defaultValue
}
