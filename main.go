package backlog

import (
	"log"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("deep-mcp: config error: %v", err)
	}

	s := buildServer(cfg)

	switch transport() {
	case "http":
		if err := runHTTP(s, httpAddr()); err != nil {
			log.Fatalf("deep-mcp: http server error: %v", err)
		}
	default:
		if err := runStdio(s); err != nil {
			log.Fatalf("deep-mcp: stdio server error: %v", err)
		}
	}
}
