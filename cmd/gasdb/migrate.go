package main

import (
	"context"
	"log/slog"

	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/urfave/cli/v2"
)

func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Migrate the fuel price database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db",
				Usage:    "Database file",
				Required: false,
				Value:    "fuel_prices.db",
			},
		},
		Action: migrateAction,
	}
}

func migrateAction(c *cli.Context) error {
	ctx := context.Background()
	_, err := gasdb.NewStorageMigrate(ctx, c.String("db"), slog.New(slog.DiscardHandler))
	return err
}
