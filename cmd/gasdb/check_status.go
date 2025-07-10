package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/urfave/cli/v2"
)

func checkStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "check-status",
		Usage: "Check for days with missing fuel prices",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db",
				Usage:    "Database file",
				Required: false,
				Value:    "fuel_prices.db",
			},
			&cli.StringFlag{
				Name:     "start",
				Usage:    "Start date (YYYY-MM-DD)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "end",
				Usage:    "End date (YYYY-MM-DD)",
				Required: false,
			},
		},
		Action: checkStatusAction,
	}
}

func checkStatusAction(c *cli.Context) error {
	ctx := context.Background()
	dbPath := c.String("db")
	storage, err := gasdb.NewStorage(ctx, dbPath, slog.New(slog.DiscardHandler))
	if err != nil {
		return err
	}
	defer storage.Close()

	allDates, err := storage.GetAllDates(ctx)
	if err != nil {
		return err
	}
	if len(allDates) == 0 {
		fmt.Println("No dates found in database.")
		return nil
	}

	var startDate, endDate time.Time
	if c.String("start") != "" {
		startDate, err = time.Parse("2006-01-02", c.String("start"))
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
	} else {
		startDate = time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if c.String("end") != "" {
		endDate, err = time.Parse("2006-01-02", c.String("end"))
		if err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}
	} else {
		endDate = time.Now()
	}

	fmt.Printf("Checking for missing days in range: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	dateSet := make(map[string]struct{}, len(allDates))
	for _, d := range allDates {
		dateSet[d.Format("2006-01-02")] = struct{}{}
	}

	var missing []string
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		if _, ok := dateSet[ds]; !ok {
			missing = append(missing, ds)
		}
	}

	if len(missing) == 0 {
		fmt.Println("No missing days in the given range.")
	} else {
		fmt.Println("Missing days:")
		for _, m := range missing {
			fmt.Println(m)
		}
	}
	return nil
}
