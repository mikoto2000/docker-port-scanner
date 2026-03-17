package dockercli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/example/docker-port-scanner/internal/app"
)

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type Scanner struct {
	runner commandRunner
}

func New() *Scanner {
	return &Scanner{runner: execRunner{}}
}

func (s *Scanner) ListRunning(ctx context.Context) ([]app.Container, error) {
	output, err := s.runner.Run(ctx, "docker", "ps", "--format", "{{.ID}}")
	if err != nil {
		return nil, wrapCommandError(err)
	}

	lines := nonEmptyLines(string(output))
	containers := make([]app.Container, 0, len(lines))
	for _, id := range lines {
		container, err := s.GetContainer(ctx, id)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})
	return containers, nil
}

func (s *Scanner) GetContainer(ctx context.Context, reference string) (app.Container, error) {
	output, err := s.runner.Run(ctx, "docker", "inspect", reference)
	if err != nil {
		return app.Container{}, wrapInspectError(reference, err)
	}

	var inspected []inspectContainer
	if err := json.Unmarshal(output, &inspected); err != nil {
		return app.Container{}, fmt.Errorf("failed to decode docker inspect output for %q: %w", reference, err)
	}
	if len(inspected) == 0 {
		return app.Container{}, fmt.Errorf("container %q was not found", reference)
	}

	container := inspected[0]
	if !container.State.Running {
		return app.Container{}, fmt.Errorf("container %q is not running", reference)
	}

	return app.Container{
		Name:  strings.TrimPrefix(container.Name, "/"),
		Ports: extractPublishedTCPPorts(container.NetworkSettings.Ports),
	}, nil
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return output, nil
}

type inspectContainer struct {
	Name            string `json:"Name"`
	State           inspectState
	NetworkSettings inspectNetworkSettings
}

type inspectState struct {
	Running bool `json:"Running"`
}

type inspectNetworkSettings struct {
	Ports map[string][]inspectPortBinding `json:"Ports"`
}

type inspectPortBinding struct {
	HostPort string `json:"HostPort"`
}

func extractPublishedTCPPorts(ports map[string][]inspectPortBinding) []int {
	if len(ports) == 0 {
		return nil
	}

	result := make([]int, 0, len(ports))
	for portDef, bindings := range ports {
		if !strings.HasSuffix(portDef, "/tcp") || len(bindings) == 0 {
			continue
		}
		for _, binding := range bindings {
			var port int
			if _, err := fmt.Sscanf(binding.HostPort, "%d", &port); err == nil {
				result = append(result, port)
			}
		}
	}

	sort.Ints(result)
	return deduplicateInts(result)
}

func deduplicateInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}

	result := []int{values[0]}
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func nonEmptyLines(text string) []string {
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func wrapInspectError(reference string, err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "No such object"):
		return fmt.Errorf("container %q was not found", reference)
	case strings.Contains(message, "permission denied"):
		return errors.New("failed to connect to Docker daemon: permission denied")
	default:
		return fmt.Errorf("failed to inspect container %q: %w", reference, err)
	}
}

func wrapCommandError(err error) error {
	if strings.Contains(err.Error(), "permission denied") {
		return errors.New("failed to connect to Docker daemon: permission denied")
	}
	return fmt.Errorf("failed to list running containers: %w", err)
}
