package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func BuildURL(scheme, host string, port int) string {
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

func RenderOutput(containers []Container, scheme, host string) (string, error) {
	if len(containers) == 0 {
		return "", errors.New("no containers to render")
	}

	var b strings.Builder
	for i, container := range containers {
		if len(container.Ports) == 0 {
			return "", fmt.Errorf("container %q has no published TCP ports", container.Name)
		}

		ports := append([]int(nil), container.Ports...)
		sort.Ints(ports)

		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(container.Name)
		b.WriteString("\n")
		for _, port := range ports {
			b.WriteString(BuildURL(scheme, host, port))
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}
