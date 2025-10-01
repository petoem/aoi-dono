package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adrg/xdg"
	"github.com/go-ini/ini"
)

var defaultConfigFileName = "aoi-dono/aoi-dono.ini"

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
	fmt.Printf("config file saved to: %s\n", file)
	return nil
}

// flagsAndEnv sets the flags and env values on the config.
//
// 1. Commandline flags, 2. environment variables and 3. config file.
func (c *config) flagsAndEnv(flags *flag.FlagSet) *bool {
	// redact credentials from usage text
	flags.SetOutput(newRedactedWriter(os.Stderr, c.Mastodon.AccessToken, c.Mastodon.ClientSecret, c.Bluesky.Password))
	// credentials
	// - Mastodon
	flags.StringVar(&c.Mastodon.Server, "mastodon-instance-url", osEnvOrConfigValue("MASTODON_INSTANCE_URL", c.Mastodon.Server), "Mastodon instance URL (e.g., https://mastodon.example)")
	flags.StringVar(&c.Mastodon.AccessToken, "mastodon-access-token", osEnvOrConfigValue("MASTODON_ACCESS_TOKEN", c.Mastodon.AccessToken), "Mastodon access token")
	flags.StringVar(&c.Mastodon.ClientID, "mastodon-client-key", osEnvOrConfigValue("MASTODON_CLIENT_KEY", c.Mastodon.ClientID), "Mastodon client key")
	flags.StringVar(&c.Mastodon.ClientSecret, "mastodon-client-secret", osEnvOrConfigValue("MASTODON_CLIENT_SECRET", c.Mastodon.ClientSecret), "Mastodon client secret")
	// - Bluesky
	flags.StringVar(&c.Bluesky.ServiceUrl, "bluesky-service-url", osEnvOrConfigValue("BLUESKY_SERVICE_URL", c.Bluesky.ServiceUrl), "Bluesky service URL (e.g., https://bsky.social)")
	flags.StringVar(&c.Bluesky.Identifier, "bluesky-identifier", osEnvOrConfigValue("BLUESKY_IDENTIFIER", c.Bluesky.Identifier), "Bluesky identifier (e.g., @user.bsky.social)")
	flags.StringVar(&c.Bluesky.Password, "bluesky-password", osEnvOrConfigValue("BLUESKY_PASSWORD", c.Bluesky.Password), "Bluesky password")

	// other flags
	df := "en"
	if c.DefaultLanguage != "" {
		df = c.DefaultLanguage
	}
	flags.StringVar(&c.DefaultLanguage, "lang", df, "Post language (e.g., jp)")

	shouldSave := flags.Bool("save-to-config", false, "Save current config to file")
	return shouldSave
}

func (m Mastodon) IsEmpty() bool {
	return m == Mastodon{}
}

func (b Bluesky) IsEmpty() bool {
	return b == Bluesky{}
}

func osEnvOrConfigValue(env, defaultValue string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	}
	return defaultValue
}

type redactedWriter struct {
	io.Writer
	redact []string
}

func newRedactedWriter(writer io.Writer, redact ...string) *redactedWriter {
	return &redactedWriter{Writer: writer, redact: redact}
}

func (r redactedWriter) Write(p []byte) (int, error) {
	tmp := make([]byte, len(p))
	copy(tmp, p)
	for _, r := range r.redact {
		if r == "" {
			continue
		}
		tmp = bytes.ReplaceAll(tmp, []byte(r), bytes.Repeat([]byte{byte('*')}, len(r)))
	}
	return r.Writer.Write(tmp)
}
