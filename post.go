package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/rivo/uniseg"
	"mvdan.cc/xurls/v2"
)

func createPost() (string, error) {
	post, err := os.CreateTemp("", "post.*.txt")
	if err != nil {
		return "", fmt.Errorf("could not create post file: %w", err)
	}
	defer post.Close()
	return post.Name(), nil
}

func readPost(post string) string {
	content, err := os.ReadFile(post)
	if err != nil {
		return ""
	}
	return string(content)
}

func splitPostIntoThread(post string, limit int, endmarker string) ([]string, error) {
	thread := make([]string, 0)
	nocutpointzones := findAllLinksInPost(post)
	limit -= uniseg.GraphemeClusterCount(endmarker)
	if limit <= 0 {
		return nil, errors.New("endmarker is bigger or equal to post character limit")
	}

	g := uniseg.NewGraphemes(post)
	count := 0
	lastSentence := struct{ end, count int }{-1, 0}
	lastLineCanBreak := struct{ end, count int }{-1, 0}
	lastcutpoint := 0
	for g.Next() {
		count++
		if count == limit {
			cutpoint := -1  // cutpoint to use as thread part end
			finalcount := 0 // final grapheme cluster count
			switch {
			case lastSentence.end != -1:
				cutpoint = lastSentence.end
				finalcount = lastSentence.count
				lastSentence.end = -1
			case lastLineCanBreak.end != -1:
				cutpoint = lastLineCanBreak.end
				finalcount = lastLineCanBreak.count
				lastLineCanBreak.end = -1
			default:
				_, cutpoint = g.Positions()
				finalcount = count
			}

			thread = append(thread, fmt.Sprintf("%s%s", post[lastcutpoint:cutpoint], endmarker))
			lastcutpoint = cutpoint
			count = count - finalcount
		}
		// skip this grapheme cluster
		if slices.ContainsFunc(nocutpointzones, func(zone []int) bool {
			b, e := g.Positions()
			return zone[0] <= b && e <= zone[1]
		}) {
			continue
		}
		// look for a sentence boundary and line we can break on, in the remaining X% of the available post length
		if 0.90 <= float64(count)/float64(limit) && g.IsSentenceBoundary() {
			_, lastSentence.end = g.Positions()
			lastSentence.count = count
		}
		if 0.95 <= float64(count)/float64(limit) && g.LineBreak() == uniseg.LineCanBreak {
			_, lastLineCanBreak.end = g.Positions()
			lastLineCanBreak.count = count
		}
	}

	if count > 0 {
		thread = append(thread, post[lastcutpoint:])
	}
	return thread, nil
}

func findAllLinksInPost(text string) [][]int {
	ru := xurls.Relaxed()
	locs := ru.FindAllStringIndex(text, -1)
	return locs
}
