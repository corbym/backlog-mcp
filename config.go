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
	if err != nil {
		return nil, fmt.Errorf("requirements directory %q does not exist: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("requirements directory %q is not a directory", root)
	}
	return &Config{StoriesRoot: root}, nil
}
