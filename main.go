package main

import (
	"os"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/internal/server"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: config.C.Service.Name,
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(c *cli.Context) error {
					s := server.New()
					s.RegisterRoutes()
					config.RegisterShutdown("server", s.Shutdown)
					return s.Listen(config.C.Service.Addr)
				},
			},
		},
	}

	go func() {
		if err := app.Run(os.Args); err != nil {
			panic(err)
		}
	}()

	config.WaitForExit()
}
