package main

import (
	"os"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/internal/server"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: config.C.ServiceName,
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(c *cli.Context) error {
					s := server.New()
					s.RegisterFiberRoutes()
					config.RegisterShutdown("fiber-server", s.Shutdown)
					return s.Listen(":8080")
				},
			},
		},
	}

	go func() {
		if err := app.Run(os.Args); err != nil {
			panic(err)
		}
	}()

	config.Shutdown()
}
