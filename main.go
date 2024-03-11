package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/internal/server"
	"github.com/tlipoca9/asta/pkg/logx"
)

func main() {
	log := slog.Default()

	app := &cli.App{
		Name: config.C.Service.Name,
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(_ *cli.Context) error {
					return server.Serve()
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if err := app.Run(os.Args); err != nil {
			log.Error("app run failed", logx.JSON("error", err))
		}
	}()

	config.WaitForExit(ctx)
}
