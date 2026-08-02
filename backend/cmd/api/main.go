package main

import (
	"flag"
	"log"

	"github.com/Dovud1997/Dovud/backend/internal/app"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	application, err := app.New(*cfgPath)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
