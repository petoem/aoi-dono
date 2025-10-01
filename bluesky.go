package main

import (
	"context"
	"errors"
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

	thread, err := splitPostIntoThread(postContent, 300, "...")
	if err != nil {
		return "", err
	}

	records := make([]*bsky.FeedPost, 0, len(thread))
	for _, p := range thread {
		facetsLink := findRichtextFacetLinks(p)
		facetsTag := findRichtextFacetTags(p)
		facetsMention := findRichtextFacetMention(ctx, c, p)

		records = append(records, &bsky.FeedPost{
			CreatedAt: string(syntax.DatetimeNow()),
			Langs:     []string{language},
			Facets:    slices.Concat(facetsLink, facetsTag, facetsMention),
			Text:      p,
		})
	}

	responses := make([]*atproto.RepoCreateRecord_Output, 0, len(thread))
	for i, record := range records {
		if i > 0 {
			record.Reply = &bsky.FeedPost_ReplyRef{
				Root: &atproto.RepoStrongRef{
					Cid: responses[0].Cid,
					Uri: responses[0].Uri,
				},
				Parent: &atproto.RepoStrongRef{
					Cid: responses[i-1].Cid,
					Uri: responses[i-1].Uri,
				},
			}
		}
		request := &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       c.AccountDID.String(),
			Record:     &util.LexiconTypeDecoder{Val: record},
		}

		res, err := atproto.RepoCreateRecord(ctx, c, request)
		if err != nil {
			return "", fmt.Errorf("create record failed: %w", err)
		}
		responses = append(responses, res)
	}

	if len(responses) == 0 {
		return "", errors.New("no record was created")
	}

	uri, err := syntax.ParseATURI(responses[0].Uri) // root post uri
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", uri.Authority(), uri.RecordKey()), nil
}

func findRichtextFacetLinks(text string) []*bsky.RichtextFacet {
	facets := make([]*bsky.RichtextFacet, 0)
	locs := findAllLinksInPost(text)
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
