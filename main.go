package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Config precedence
// 1. Commandline args
// 2. Environment Variable
// 3. Config file
func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// load config file
	cfg := new(config)
	err := cfg.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// commandline flags
	// - Platform independent
	language := flag.String("lang", "en", "Post language (e.g., jp)")
	// - Mastodon
	mastodonInstanceUrl := flag.String("mastodonInstanceUrl", os.Getenv("MASTODON_INSTANCE_URL"), "Mastodon instance URL (e.g., https://mastodon.example)")
	mastodonAccessToken := flag.String("mastodonAccessToken", os.Getenv("MASTODON_ACCESS_TOKEN"), "Mastodon access token")
	mastodonClientKey := flag.String("mastodonClientKey", os.Getenv("MASTODON_CLIENT_KEY"), "Mastodon client key")
	mastodonClientSecret := flag.String("mastodonClientSecret", os.Getenv("MASTODON_CLIENT_SECRET"), "Mastodon client secret")
	// - Bluesky
	blueskyServiceUrl := flag.String("blueskyServiceUrl", os.Getenv("BLUESKY_SERVICE_URL"), "Bluesky service URL (e.g., https://bsky.social)")
	blueskyIdentifier := flag.String("blueskyIdentifier", os.Getenv("BLUESKY_IDENTIFIER"), "Bluesky identifier (e.g., @user.bsky.social)")
	blueskyPassword := flag.String("blueskyPassword", os.Getenv("BLUESKY_PASSWORD"), "Bluesky password")
	// - misc
	saveToConfigFile := flag.Bool("saveToConfig", false, "Save commandline flags to config file")
	flag.Parse()

	// save config to file
	if *saveToConfigFile {
		err := cfg.SaveConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	editor := getEditor()
	post, err := createPost()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.Remove(post)
	err = openEditor(editor, post)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	postContent := readPost(post)

	// TODO: detect if post is longer than Bluesky 300 characters
	// limit and split it into a thread
	url, err := blueskyPost(ctx,
		*blueskyServiceUrl,
		*blueskyIdentifier,
		*blueskyPassword,
		*language,
		postContent,
	)
	if err != nil {
		fmt.Printf("\x1b[1;31m✗\x1b[0m %s: %s\n", "Bluesky", err)
	} else {
		fmt.Printf("\x1b[1;32m✓\x1b[0m %s: %s\n", "Bluesky", url)
	}

	url, err = mastodonPost(ctx,
		*mastodonInstanceUrl,
		*mastodonAccessToken,
		*mastodonClientKey,
		*mastodonClientSecret,
		*language,
		postContent,
	)
	if err != nil {
		fmt.Printf("\x1b[1;31m✗\x1b[0m %s: %s\n", "Mastodon", err)
	} else {
		fmt.Printf("\x1b[1;32m✓\x1b[0m %s: %s\n", "Mastodon", url)
	}
}
