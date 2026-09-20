package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"summary-your-usage/config"
	"summary-your-usage/migrations"
	"summary-your-usage/pkg/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()
	command := flag.Arg(0)
	if flag.NArg() != 1 || (command != "up" && command != "down" && command != "status") {
		return fmt.Errorf("usage: go run ./cmd/migrate [-config path] [up|down|status]")
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	provider, err := migrations.New(db, cfg.Database.Driver)
	if err != nil {
		return err
	}
	switch command {
	case "up":
		_, err = provider.Up(ctx)
	case "down":
		_, err = provider.Down(ctx)
	case "status":
		statuses, statusErr := provider.Status(ctx)
		if statusErr != nil {
			return statusErr
		}
		for _, status := range statuses {
			fmt.Printf("%s\t%s\n", status.State, status.Source.Path)
		}
	}
	return err
}
