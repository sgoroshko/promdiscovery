package cmd

import (
	"context"
	"fmt"
	"sort"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// NewCommandMerge prepare services
func NewCommandCompose() *cli.Command {
	cfg := new(configCompose)
	return &cli.Command{
		Name:   "compose",
		Usage:  "",
		Flags:  bindConfigCompose(cfg),
		Action: actionCompose(cfg),
	}
}

type configCompose struct {
	debug      bool
	dockerHost string
	network    string
	key        string
	filename   string
}

func bindConfigCompose(cfg *configCompose) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug mode",
			Destination: &cfg.debug,
			Value:       false,
		},
		&cli.StringFlag{
			Name:        "dockerHost",
			Usage:       "docker host",
			Destination: &cfg.dockerHost,
			Value:       "unix:///var/run/docker.sock",
		},
		&cli.StringFlag{
			Name:        "network",
			Usage:       "docker network, optional",
			Destination: &cfg.network,
		},
		&cli.StringFlag{
			Name:        "key",
			Usage:       "scrape key",
			Destination: &cfg.key,
			Value:       "metrics",
		},
		&cli.StringFlag{
			Name:        "output",
			Usage:       "output filename",
			Destination: &cfg.filename,
			Value:       "discovered.json",
		},
	}
}

func actionCompose(cfg *configCompose) cli.ActionFunc {
	return func(c *cli.Context) error {
		logrus.Infof("start %s with config: %+v", c.App.Name, *cfg)

		if cfg.debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		ctx := cancelOnSignal(context.Background(),
			syscall.SIGINT,
			syscall.SIGTERM)

		return observeCompose(ctx, cfg)
	}
}

func observeCompose(ctx context.Context, cfg *configCompose) error {
	dockerClient, err := client.NewClient(cfg.dockerHost, client.DefaultVersion, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	const timeout = 12 * time.Second
	scrapeRunner := make(chan struct{}, 1) //
	scrapeTrigger := true
	go func() { // run scrape on start
		time.Sleep(timeout)
		scrapeRunner <- struct{}{}
	}()

	args := filters.NewArgs()
	args.Add("type", events.ContainerEventType)
	eventsCh, errorsCh := dockerClient.Events(ctx, types.EventsOptions{Filters: args})
	for {
		select {
		case ev := <-eventsCh:
			if ev.Action == "start" || ev.Action == "die" {
				logrus.Debugf("dockerd event: %-10s %-10s %s", ev.Type, ev.Action, ev.ID)
				if !scrapeTrigger {
					scrapeTrigger = true
					go func() {
						time.Sleep(timeout)
						scrapeRunner <- struct{}{}
					}()
				}
			}

		case <-scrapeRunner:
			scrapeTrigger = false
			err := scrapeTargetsCompose(ctx, cfg, dockerClient)
			if err != nil {
				return errors.WithStack(err)
			}

		case err := <-errorsCh:
			return errors.WithStack(err)
		}
	}
}

func scrapeTargetsCompose(ctx context.Context, cfg *configCompose, dockerClient *client.Client) error {
	logrus.Debugf("scrapping targets...")

	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return errors.WithStack(err)
	}

	type prometheusTargets struct {
		Targets []string          `json:"targets,omitempty"`
		Labels  map[string]string `json:"labels,omitempty"`
	}

	scrappedTargets := make([]*prometheusTargets, 0, len(containers))
	for _, c := range containers {
		if cfg.network != "" && !containerHasNetwork(c, cfg.network) {
			logrus.Debugf("container %s no have network: %s", c.ID, cfg.network)
			continue
		}

		pt := new(prometheusTargets)

		for k, v := range c.Labels {
			if k == cfg.key {
				pt.Targets = append(pt.Targets, fmt.Sprintf("%s:%s", c.Names[0][1:], v))
			}
		}

		if len(pt.Targets) > 0 {
			sort.Strings(pt.Targets)
			scrappedTargets = append(scrappedTargets, pt)
		}
	}

	sort.Slice(scrappedTargets, func(i, j int) bool {
		return scrappedTargets[i].Targets[0] < scrappedTargets[j].Targets[0]
	})

	err = writeDataIntoFileIfChanged(cfg.filename, scrappedTargets)
	return errors.WithStack(err)
}

func containerHasNetwork(container types.Container, network string) bool {
	if container.NetworkSettings != nil {
		for k := range container.NetworkSettings.Networks {
			if k == network {
				return true
			}
		}
	}

	return false
}
