package main

import (
	"fmt"
	"log"
	"os"
)

// Version is set at build time via -ldflags "-X main.Version=<tag>".
var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(Version)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		args := os.Args[2:]
		dir := "requirements"
		if len(args) > 0 {
			dir = args[0]
		}
		if err := runInit(dir); err != nil {
			log.Fatalf("backlog-mcp: init error: %v", err)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "plan" {
		name := ""
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		cfg, err := LoadConfig()
		if err != nil {
			log.Fatalf("backlog-mcp: config error: %v", err)
		}
		if err := runPlan(cfg.StoriesRoot, name); err != nil {
			log.Fatalf("backlog-mcp: plan error: %v", err)
		}
		return
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("backlog-mcp: config error: %v", err)
	}

	s := buildServer(cfg)

	if err := runStdio(s); err != nil {
		log.Fatalf("backlog-mcp: stdio server error: %v", err)
	}
}
