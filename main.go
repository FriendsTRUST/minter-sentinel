package main

import (
	"minter-sentinel/cmd/start"
	"minter-sentinel/cmd/txgenerate"
	"minter-sentinel/config"
	"os"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	log := logrus.New()

	logrus.SetReportCaller(true)

	var cfg config.Config

	startCmd := start.New(log, &cfg)
	txGenerateCmd := txgenerate.New(log, &cfg)

	app := &cli.App{
		Name:     "minter-sentinel",
		Version:  "1.0.0",
		Usage:    "Watch for missed blocks",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Vladimir Yuldashev",
				Email: "misterio92@gmail.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				Value:   "config.yaml",
			},
		},
		Before: func(ctx *cli.Context) error {
			c, err := config.New(ctx.String("config"))

			if err != nil {
				return err
			}

			cfg = *c

			return nil
		},
		Commands: []*cli.Command{
			startCmd.Command(),
			txGenerateCmd.Command(),
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
