package main

import (
	"os"
	"sort"

	"github.com/disaster37/crossplane-gen/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func run(args []string) error {

	// Logger setting
	log.SetOutput(os.Stdout)

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Crossplane CRD generator from golang base on operator-sdk"
	app.Version = "develop"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Display debug output",
		},
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "No print color",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:     "crd",
			Usage:    "Generate CRD",
			Category: "CRD",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "source-path",
					Usage:    "The source path from where generate CRD",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "target-path",
					Usage:    "The target path to where generate CRD",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:     "crd-options",
					Usage:    "CRD option to pass on controller-gen",
					Required: false,
				},
				&cli.StringSliceFlag{
					Name:     "schemapatch-options",
					Usage:    "CRD option to pass on controller-gen",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "claim-name",
					Usage:    "The claim name",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "claim-plural-name",
					Usage:    "The claim name plural",
					Required: false,
				},
			},
			Action: cmd.GenerateCRD,
		},
	}

	app.Before = func(c *cli.Context) error {

		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if !c.Bool("no-color") {
			formatter := new(prefixed.TextFormatter)
			formatter.FullTimestamp = true
			formatter.ForceFormatting = true
			log.SetFormatter(formatter)
		}

		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(args)
	return err
}

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
