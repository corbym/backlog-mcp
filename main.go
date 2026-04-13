package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		args := os.Args[2:]
		if len(args) == 0 {
			log.Fatal("backlog-mcp: usage: backlog-mcp init <directory>")
		}
		if err := runInit(args[0]); err != nil {
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

	switch transport() {
	case "http":
		if err := runHTTP(s, httpAddr()); err != nil {
			log.Fatalf("backlog-mcp: http server error: %v", err)
		}
	default:
		if err := runStdio(s); err != nil {
			log.Fatalf("backlog-mcp: stdio server error: %v", err)
		}
	}
}
