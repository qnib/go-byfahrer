package main

import (
	"log"
	"os"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-collector-docker-events/lib"
	"github.com/qnib/qframe-types"

	"github.com/qnib/go-byfahrer/lib"
)

func check_err(pname string, err error) {
	if err != nil {
		log.Printf("[EE] Failed to create %s plugin: %s", pname, err.Error())
		os.Exit(1)
	}
}

func Run(ctx *cli.Context) {
	// Create conf
	log.Printf("[II] Start Version: %s", ctx.App.Version)
	cfg := config.NewConfig([]config.Provider{config.NewCLI(ctx, true)})
	cfg.Providers = append(cfg.Providers, )
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	gb, err := byfahrer.New(qChan, cfg, "go-byfahrer")
	check_err(gb.Name, err)
	go gb.Run()
	pe, err := qframe_collector_docker_events.New(qChan, cfg, "docker-events")
	check_err(pe.Name, err)
	go pe.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = "Start container to terminate SSL for others."
	app.Usage = "go-byfahrer [options]"
	app.Version = "0.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-host",
			Value: "unix:///var/run/docker.sock",
			Usage: "Docker host to connect to.",
			EnvVar: "DOCKER_HOST",
		},
	}
	app.Action = Run
	app.Run(os.Args)
}
