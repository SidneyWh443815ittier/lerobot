// Package main is the entry point for lerobot, a bot that automates
// GitHub repository maintenance tasks for the Flatcar Linux project.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/flatcar/lerobot/pkg/bot"
	"github.com/flatcar/lerobot/pkg/config"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "unknown"
)

func main() {
	var (
		configPath  = flag.String("config", "config.yaml", "path to configuration file")
		showVersion = flag.Bool("version", false, "print version information and exit")
		dryRun      = flag.Bool("dry-run", false, "run without making any changes")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("lerobot version %s (commit: %s)\n", Version, Commit)
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting lerobot %s (commit: %s)", Version, Commit)

	// Load configuration from file.
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if *dryRun {
		log.Println("Dry-run mode enabled: no changes will be made")
		cfg.DryRun = true
	}

	// Create a cancellable context that responds to OS signals.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on SIGINT or SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %s, shutting down...", sig)
		cancel()
	}()

	// Initialise and run the bot.
	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialise bot: %v", err)
	}

	if err := b.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Bot exited with error: %v", err)
	}

	log.Println("lerobot stopped cleanly")
}
