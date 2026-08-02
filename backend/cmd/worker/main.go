package main

import (
	"flag"
	"log"

	"github.com/Dovud1997/Dovud/backend/internal/worker"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	runner, err := worker.New(*cfgPath)
	if err != nil {
		log.Fatalf("worker bootstrap failed: %v", err)
	}
	if err := runner.Run(); err != nil {
		log.Fatalf("worker stopped: %v", err)
	}
}
