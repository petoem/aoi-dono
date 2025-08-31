package main

import (
	"fmt"
	"os"
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
