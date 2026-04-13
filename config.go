package main

import (
	"fmt"
	"os"
)

type Config struct {
	StoriesRoot string
}

func LoadConfig() (*Config, error) {
	root := os.Getenv("BACKLOG_ROOT")
	if root == "" {
		root = "requirements"
	}
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		if mkErr := os.MkdirAll(root, 0755); mkErr != nil {
			return nil, fmt.Errorf("could not create requirements directory %q: %w", root, mkErr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("requirements directory %q: %w", root, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("requirements directory %q is not a directory", root)
	}
	return &Config{StoriesRoot: root}, nil
}
