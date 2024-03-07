package main

import (
	"context"
	"github.com/tlipoca9/asta/internal/config"
	"os"

	"github.com/tlipoca9/asta/internal/server"
	"github.com/urfave/cli/v2"
)

func main() {
	defer config.Shutdown(context.Background())

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
