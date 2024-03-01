package main

import (
	"log/slog"
	"os"

	"github.com/tlipoca9/asta/internal/server"
	"github.com/urfave/cli/v2"
)

func init() {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
}

func main() {

	app := &cli.App{
		Name: "asta",
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(c *cli.Context) error {
					s := server.New()
					s.RegisterFiberRoutes()
					return s.Listen(":8080")
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
