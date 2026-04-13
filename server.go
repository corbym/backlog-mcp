package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

func runHTTP(s *server.MCPServer, addr string) error {
	log.Printf("backlog-mcp: starting HTTP/SSE server on %s", addr)
	sseServer := server.NewSSEServer(s,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
	)
	return http.ListenAndServe(addr, sseServer)
}

// transport returns "stdio" or "http" based on the BACKLOG_TRANSPORT env var.
// Default is stdio.
func transport() string {
	t := os.Getenv("BACKLOG_TRANSPORT")
	if t == "http" {
		return "http"
	}
	return "stdio"
}

// httpAddr returns the address to listen on for HTTP mode.
// Default: 0.0.0.0:8080, overridable via BACKLOG_HTTP_ADDR.
func httpAddr() string {
	addr := os.Getenv("BACKLOG_HTTP_ADDR")
	if addr == "" {
		return "0.0.0.0:8080"
	}
	return addr
}
