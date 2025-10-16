package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
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

	// commandline flags
	shouldSave := cfg.flagsAndEnv(flag.CommandLine)
	printversion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// print version information
	if *printversion {
		fmt.Printf("%s %s\n", filepath.Base(os.Args[0]), version())
		os.Exit(0)
	}

	// save config to file
	if *shouldSave {
		err := cfg.SaveConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	post, err := NewPost(cfg.DefaultLanguage)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer post.Delete()
	editor := getEditor()
	err = openEditor(editor, post.Filepath())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if strings.TrimSpace(post.Content()) == "" {
		fmt.Println("Nothing to post, exiting.")
		return
	}

	if !cfg.Bluesky.IsEmpty() {
		url, err := blueskyPost(ctx,
			cfg.Bluesky,
			post,
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
			post,
		)
		if err != nil {
			fmt.Printf("\x1b[1;31m✗\x1b[0m %s: %s\n", "Mastodon", err)
		} else {
			fmt.Printf("\x1b[1;32m✓\x1b[0m %s: %s\n", "Mastodon", url)
		}
	}
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}

	return "unknown"
}
