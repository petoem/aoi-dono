package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattn/go-mastodon"
)

const defaultMastodonStatusCharacterLimit = 500

func mastodonPost(ctx context.Context, credentials Mastodon, post *Post) (string, error) {
	config := &mastodon.Config{
		Server:       credentials.Server,
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		AccessToken:  credentials.AccessToken,
	}
	client := mastodon.NewClient(config)

	instance, err := client.GetInstance(ctx)
	if err != nil {
		return "", err
	}
	characterlimit, ok := (*instance.GetConfig().Statuses)["max_characters"].(float64)
	if !ok {
		characterlimit = defaultMastodonStatusCharacterLimit
	}

	thread, err := post.SplitIntoThread(int(characterlimit), "...")
	if err != nil {
		return "", err
	}

	toots := make([]*mastodon.Toot, 0, len(thread))
	for _, p := range thread {
		toots = append(toots, &mastodon.Toot{
			Status:     p.Content(),
			Visibility: "public",
			Language:   p.Language(),
		})
	}

	statuses := make([]*mastodon.Status, 0, len(toots))
	for i, toot := range toots {
		if i > 0 {
			toot.InReplyToID = statuses[i-1].ID // set current toot to be a reply to the previous one
		}
		status, err := mastodonPostStatus(ctx, client, toot)
		if err != nil {
			return "", fmt.Errorf("error sending %d. toot in thread of length %d: %w", i+1, len(toots), err)
		}
		statuses = append(statuses, status)
	}

	if len(statuses) == 0 {
		return "", errors.New("no toots where send")
	}

	return statuses[0].URL, nil
}

func mastodonPostStatus(ctx context.Context, client *mastodon.Client, toot *mastodon.Toot) (*mastodon.Status, error) {
	status, err := client.PostStatus(ctx, toot)
	if err != nil {
		return nil, fmt.Errorf("failed to post status: %w", err)
	}
	return status, nil
}
