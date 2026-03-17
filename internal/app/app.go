package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

type Container struct {
	Name  string
	Ports []int
}

type Scanner interface {
	ListRunning(ctx context.Context) ([]Container, error)
	GetContainer(ctx context.Context, reference string) (Container, error)
}

type Config struct {
	Host      string
	Scheme    string
	Targets   []string
	ShowUsage func()
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, scanner Scanner) error {
	cfg, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}

	containers, err := collectContainers(ctx, scanner, cfg.Targets)
	if err != nil {
		return err
	}

	output, err := RenderOutput(containers, cfg.Scheme, cfg.Host)
	if err != nil {
		return err
	}

	_, err = io.WriteString(stdout, output)
	return err
}

func parseArgs(args []string, stderr io.Writer) (Config, error) {
	cfg := Config{
		Host:   "localhost",
		Scheme: "http",
	}

	fs := flag.NewFlagSet("docker-port-scanner", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&cfg.Host, "host", cfg.Host, "hostname part of generated URLs")
	fs.StringVar(&cfg.Scheme, "scheme", cfg.Scheme, "scheme part of generated URLs")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.Targets = fs.Args()
	return cfg, nil
}

func collectContainers(ctx context.Context, scanner Scanner, targets []string) ([]Container, error) {
	if len(targets) > 0 {
		containers := make([]Container, 0, len(targets))
		for _, target := range targets {
			container, err := scanner.GetContainer(ctx, target)
			if err != nil {
				return nil, err
			}
			if len(container.Ports) == 0 {
				return nil, fmt.Errorf("container %q has no published TCP ports", container.Name)
			}
			containers = append(containers, container)
		}
		return containers, nil
	}

	containers, err := scanner.ListRunning(ctx)
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, errors.New("no running containers found")
	}

	filtered := make([]Container, 0, len(containers))
	for _, container := range containers {
		if len(container.Ports) == 0 {
			continue
		}
		filtered = append(filtered, container)
	}
	if len(filtered) == 0 {
		return nil, errors.New("no running containers expose published TCP ports")
	}

	return filtered, nil
}
