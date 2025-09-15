package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// load config file
	cfg := new(config)
	err := cfg.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	shouldSave := cfg.parseFlagsAndEnv()

	// save config to file
	if shouldSave {
		err := cfg.SaveConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
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

	if !cfg.Bluesky.IsEmpty() {
		// TODO: detect if post is longer than Bluesky 300 characters
		// limit and split it into a thread
		url, err := blueskyPost(ctx,
			cfg.Bluesky,
			cfg.DefaultLanguage,
			postContent,
		)
		if err != nil {
			fmt.Printf("\x1b[1;31m✗\x1b[0m %s: %s\n", "Bluesky", err)
		} else {
			fmt.Printf("\x1b[1;32m✓\x1b[0m %s: %s\n", "Bluesky", url)
		}
	}

	if !cfg.Mastodon.IsEmpty() {
		url, err := mastodonPost(ctx,
			cfg.Mastodon,
			cfg.DefaultLanguage,
			postContent,
		)
		if err != nil {
			fmt.Printf("\x1b[1;31m✗\x1b[0m %s: %s\n", "Mastodon", err)
		} else {
			fmt.Printf("\x1b[1;32m✓\x1b[0m %s: %s\n", "Mastodon", url)
		}
	}
}
