package main

import (
	"log/slog"

	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/urfave/cli/v2"
)

func updateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update the fuel price database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db",
				Usage:    "Database file",
				Required: false,
				Value:    "fuel_prices.db",
			},
		},
		Action: updateAction,
	}
}

func updateAction(c *cli.Context) error {
	storage, err := gasdb.NewStorage(c.String("db"), slog.New(slog.DiscardHandler))
	if err != nil {
		return err
	}
	return storage.UpdateDBAll()
}
