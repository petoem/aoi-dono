package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/client"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
	"mvdan.cc/xurls/v2"
)

// Regex for finding mentions from: https://github.com/bluesky-social/atproto/blob/d91988fe79030b61b556dd6f16a46f0c3b9d0b44/packages/api/src/rich-text/util.ts
var mentionsRegex = regexp.MustCompile(`(^|\s|\()(@)([a-zA-Z0-9.-]+)(\b)`)

// Regex for hastags
var hastagRegex = regexp.MustCompile(`#\w+`)

func blueskyPost(ctx context.Context, credentials Bluesky, language, postContent string) (string, error) {
	c, err := client.LoginWithPasswordHost(ctx, credentials.ServiceUrl, credentials.Identifier, credentials.Password, "", nil)
	if err != nil {
		return "", err
	}

	facetsLink := findRichtextFacetLinks(postContent)
	facetsTag := findRichtextFacetTags(postContent)
	facetsMention := findRichtextFacetMention(ctx, c, postContent)

	record := bsky.FeedPost{
		CreatedAt: string(syntax.DatetimeNow()),
		Langs:     []string{language},
		Facets:    slices.Concat(facetsLink, facetsTag, facetsMention),
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

func findRichtextFacetTags(text string) []*bsky.RichtextFacet {
	facets := make([]*bsky.RichtextFacet, 0)
	locs := hastagRegex.FindAllStringIndex(text, -1)
	for _, loc := range locs {
		tag := text[loc[0]:loc[1]]
		if utf8.RuneCountInString(tag) > 64 {
			continue
		}
		facets = append(facets, &bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{
				{
					RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
						Tag: strings.TrimPrefix(tag, "#"),
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

func findRichtextFacetMention(ctx context.Context, client util.LexClient, text string) []*bsky.RichtextFacet {
	facets := make([]*bsky.RichtextFacet, 0)
	locs := mentionsRegex.FindAllStringIndex(text, -1)
	for _, loc := range locs {
		user := text[loc[0]:loc[1]]
		if t := strings.TrimLeftFunc(user, unicode.IsSpace); len(t) != len(user) { // match has a leading space
			loc[0] += len(user) - len(t)
			user = t
		}
		res, err := atproto.IdentityResolveHandle(ctx, client, strings.TrimPrefix(user, "@"))
		if err != nil {
			continue // ignore handles we can't resolve
		}
		facets = append(facets, &bsky.RichtextFacet{
			Features: []*bsky.RichtextFacet_Features_Elem{
				{
					RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
						Did: res.Did,
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
