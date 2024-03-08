package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/internal/server"
	"github.com/urfave/cli/v2"
)

func main() {
	log := slog.Default()

	app := &cli.App{
		Name: config.C.Service.Name,
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(c *cli.Context) error {
					s := server.New()
					s.RegisterRoutes()
					config.RegisterShutdown("server", s.ShutdownWithContext)
					return s.Listen(config.C.Service.Addr)
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if err := app.Run(os.Args); err != nil {
			log.Error("app run failed", "error", err)
		}
	}()

	config.WaitForExit(ctx)
}
