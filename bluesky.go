package main

import (
	"context"
	"fmt"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/client"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
)

func blueskyPost(ctx context.Context, credentials Bluesky, language, postContent string) (string, error) {
	c, err := client.LoginWithPasswordHost(ctx, credentials.ServiceUrl, credentials.Identifier, credentials.Password, "", nil)
	if err != nil {
		return "", err
	}

	record := bsky.FeedPost{
		CreatedAt: string(syntax.DatetimeNow()),
		Langs:     []string{language},
		Text:      postContent,
	}

	request := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       c.AccountDID.String(),
		Record:     &util.LexiconTypeDecoder{Val: &record},
	}

	response, err := atproto.RepoCreateRecord(ctx, c, request)
	if err != nil {
		return "", fmt.Errorf("create record failed: %w", err)
	}
	// TODO: transform into URL
	// at://<DID>/<COLLECTION>/<RKEY>
	// https://bsky.app/profile/<DID>/post/<RKEY>
	return response.Uri, nil
}
