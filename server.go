package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

func buildServer(cfg *Config) *server.MCPServer {
	s := server.NewMCPServer(
		"backlog-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)
	registerTools(s, cfg)
	return s
}

func runStdio(s *server.MCPServer) error {
	log.Println("backlog-mcp: starting in stdio mode")
	return server.ServeStdio(s)
}