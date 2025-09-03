package main

import (
	"context"
	"fmt"

	"github.com/mattn/go-mastodon"
)

func mastodonPost(ctx context.Context, credentials Mastodon, language, postContent string) (string, error) {
	config := &mastodon.Config{
		Server:       credentials.Server,
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		AccessToken:  credentials.AccessToken,
	}
	client := mastodon.NewClient(config)

	toot := mastodon.Toot{
		Status:     postContent,
		Visibility: "public",
		Language:   language,
	}

	post, err := client.PostStatus(ctx, &toot)
	if err != nil {
		return "", fmt.Errorf("failed to post toot: %w", err)
	}

	return post.URL, nil
}
