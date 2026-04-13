package backlog

import (
	"fmt"
	"os"
)

type Config struct {
	StoriesRoot string
}

func LoadConfig() (*Config, error) {
	root := os.Getenv("DEEP_STORIES_ROOT")
	if root == "" {
		return nil, fmt.Errorf("DEEP_STORIES_ROOT environment variable is not set")
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("DEEP_STORIES_ROOT %q does not exist: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("DEEP_STORIES_ROOT %q is not a directory", root)
	}
	return &Config{StoriesRoot: root}, nil
}
