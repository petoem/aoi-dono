package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/client"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
	"mvdan.cc/xurls/v2"
)

func blueskyPost(ctx context.Context, credentials Bluesky, language, postContent string) (string, error) {
	c, err := client.LoginWithPasswordHost(ctx, credentials.ServiceUrl, credentials.Identifier, credentials.Password, "", nil)
	if err != nil {
		return "", err
	}

	facetsLink := findRichtextFacetLinks(postContent)

	record := bsky.FeedPost{
		CreatedAt: string(syntax.DatetimeNow()),
		Langs:     []string{language},
		Facets:    facetsLink,
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

func findRichtextFacetLinks(text string) []*bsky.RichtextFacet {
	facets := make([]*bsky.RichtextFacet, 0)
	ru := xurls.Relaxed()
	locs := ru.FindAllStringIndex(text, -1)
	for _, loc := range locs {
		uri, err := url.Parse(text[loc[0]:loc[1]])
		if err != nil {
			continue // skip this uri
		}
		if uri.Scheme == "" {
			uri.Scheme = "https"
		}
		facets = append(facets, &bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{
				{
					RichtextFacet_Link: &bsky.RichtextFacet_Link{
						Uri: uri.String(),
					},
				},
			},
			Index: &bsky.RichtextFacet_ByteSlice{
				ByteEnd:   int64(loc[1]),
				ByteStart: int64(loc[0]),
			},
		})
	}
	return facets
}
