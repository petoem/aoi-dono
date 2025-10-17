package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/rivo/uniseg"
	"mvdan.cc/xurls/v2"
)

type Post struct {
	data     io.ReadSeeker
	language string
	// TODO: add attachment like image or video
}

func NewPostFile(language string) (*Post, error) {
	post, err := os.CreateTemp("", "post.*.txt")
	if err != nil {
		return nil, fmt.Errorf("could not create post file: %w", err)
	}
	defer post.Close()
	return &Post{
		data:     post,
		language: language,
	}, nil
}

func NewPostString(language, content string) (*Post, error) {
	reader := strings.NewReader(content)
	return &Post{
		data:     reader,
		language: language,
	}, nil
}

func (p *Post) Content() string {
	// This is needed because some editors
	// replace the original file instead of writing to it.
	// By default os.File are closed and reopened for reading.
	if file, ok := p.data.(*os.File); ok {
		f, err := os.Open(file.Name())
		if err != nil {
			panic(err)
		}
		p.data = f
		defer f.Close()
	}

	_, err := p.data.Seek(0, io.SeekStart)
	if err != nil {
		panic(fmt.Errorf("failed to seek to post data start: %w", err))
	}
	content, err := io.ReadAll(p.data)
	if err != nil {
		panic(fmt.Errorf("failed to read post data: %w", err))
	}
	return string(content)
}

func (p Post) DataPath() (path string) {
	if file, ok := p.data.(*os.File); ok {
		path = file.Name()
	}
	return path
}

func (p Post) Language() string {
	return p.language
}

func (p Post) SplitIntoThread(limit int, endmarker string) ([]*Post, error) {
	t, err := splitPostIntoThread(p.Content(), limit, endmarker)
	if err != nil {
		return nil, err
	}
	thread := make([]*Post, 0, len(t))
	for _, str := range t {
		np, _ := NewPostString(p.language, str)
		thread = append(thread, np)
	}
	return thread, nil
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
