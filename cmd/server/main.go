package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/its-the-vibe/MindingTheGap/internal/config"
	"github.com/its-the-vibe/MindingTheGap/internal/server"
	"github.com/its-the-vibe/MindingTheGap/internal/tfl"
)

func main() {
	path := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*path)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client := tfl.NewClient(cfg.TFL.BaseURL, cfg.TFL.AppKey)
	log.Printf("listening on :%d", cfg.Server.Port)
	if err := server.New(cfg, client).Run(ctx); err != nil {
		log.Print(err)
		stop()
		os.Exit(1)
	}
}
