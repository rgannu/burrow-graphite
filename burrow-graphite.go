package main

import (
	"fmt"
	"os"

  log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"context"
	"os/signal"
	"syscall"
  "github.com/rgannu/burrow-graphite/burrow_graphite"
)

var Version = "0.1"

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Name = "burrow-exporter"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "burrow-addr",
			Usage: "Address that burrow is listening on",
		},
		cli.StringFlag{
			Name:  "graphite-host",
			Usage: "Graphite host",
		},
		cli.IntFlag{
			Name:  "graphite-port",
			Usage: "The graphite port",
		},
		cli.IntFlag{
			Name:  "interval",
			Usage: "The interval(seconds) specifies how often to scrape burrow.",
		},
	}

	app.Action = func(c *cli.Context) error {
		if !c.IsSet("burrow-addr") {
			fmt.Println("A burrow address is required (e.g. --burrow-addr http://localhost:8000)")
			os.Exit(1)
		}

		if !c.IsSet("graphite-host") {
			fmt.Println("The graphite host is required (e.g. --graphite-host 192.168.99.100)")
			os.Exit(1)
		}

    var graphitePort = 2003
		if !c.IsSet("graphite-port") {
			fmt.Println("The graphite host is not set. Defaulting to 2003 (e.g. --graphite-port 2003)")
		} else {
		    graphitePort = c.Int("graphite-port")
		}

		if !c.IsSet("interval") {
			fmt.Println("A scrape interval is required (e.g. --interval 30)")
			os.Exit(1)
		}

		done := make(chan os.Signal, 1)

		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())

		exporter := burrow_graphite.MakeBurrowExporter(c.String("burrow-addr"), c.String("graphite-host"), graphitePort, c.Int("interval"))
		go exporter.Start(ctx)

		<-done
		cancel()

		exporter.Close()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error running burrow-exporter")
		os.Exit(1)
	}
}
