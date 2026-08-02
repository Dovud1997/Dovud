package main

import (
	"flag"
	"log"

	"github.com/Dovud1997/Dovud/backend/internal/scheduler"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	runner, err := scheduler.New(*cfgPath)
	if err != nil {
		log.Fatalf("scheduler bootstrap failed: %v", err)
	}
	if err := runner.Run(); err != nil {
		log.Fatalf("scheduler stopped: %v", err)
	}
}
