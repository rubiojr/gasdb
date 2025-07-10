package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gasdb",
		Usage: "Manage fuel price database and find nearby gas stations",
		Commands: []*cli.Command{
			updateCommand(),
			migrateCommand(),
			listNearbyCommand(),
			checkStatusCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
