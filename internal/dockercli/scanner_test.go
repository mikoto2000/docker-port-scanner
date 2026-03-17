package dockercli

import (
	"context"
	"errors"
	"testing"
)

type fakeRunner struct {
	run func(context.Context, string, ...string) ([]byte, error)
}

func (f fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f.run(ctx, name, args...)
}

func TestExtractPublishedTCPPortsIgnoresUDPAndDuplicates(t *testing.T) {
	ports := map[string][]inspectPortBinding{
		"80/tcp":   {{HostPort: "8080"}},
		"443/tcp":  {{HostPort: "8443"}, {HostPort: "8443"}},
		"53/udp":   {{HostPort: "5353"}},
		"8080/tcp": nil,
	}

	got := extractPublishedTCPPorts(ports)
	want := []int{8080, 8443}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestGetContainerReturnsRunningContainer(t *testing.T) {
	scanner := Scanner{
		runner: fakeRunner{
			run: func(context.Context, string, ...string) ([]byte, error) {
				return []byte(`[{
					"Name": "/web",
					"State": {"Running": true},
					"NetworkSettings": {
						"Ports": {
							"80/tcp": [{"HostPort": "8080"}],
							"443/tcp": [{"HostPort": "8443"}]
						}
					}
				}]`), nil
			},
		},
	}

	container, err := scanner.GetContainer(context.Background(), "web")
	if err != nil {
		t.Fatalf("GetContainer returned error: %v", err)
	}
	if container.Name != "web" {
		t.Fatalf("got name %q", container.Name)
	}
	wantPorts := []int{8080, 8443}
	for i := range wantPorts {
		if container.Ports[i] != wantPorts[i] {
			t.Fatalf("got ports %v, want %v", container.Ports, wantPorts)
		}
	}
}

func TestGetContainerReturnsErrorForStoppedContainer(t *testing.T) {
	scanner := Scanner{
		runner: fakeRunner{
			run: func(context.Context, string, ...string) ([]byte, error) {
				return []byte(`[{
					"Name": "/db",
					"State": {"Running": false},
					"NetworkSettings": {"Ports": {}}
				}]`), nil
			},
		},
	}

	_, err := scanner.GetContainer(context.Background(), "db")
	if err == nil || err.Error() != `container "db" is not running` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListRunningSortsByContainerName(t *testing.T) {
	scanner := Scanner{
		runner: fakeRunner{
			run: func(_ context.Context, name string, args ...string) ([]byte, error) {
				if len(args) >= 1 && args[0] == "ps" {
					return []byte("id-b\nid-a\n"), nil
				}
				if args[0] == "inspect" && args[1] == "id-b" {
					return []byte(`[{
						"Name": "/web",
						"State": {"Running": true},
						"NetworkSettings": {"Ports": {"80/tcp": [{"HostPort": "8080"}]}}
					}]`), nil
				}
				if args[0] == "inspect" && args[1] == "id-a" {
					return []byte(`[{
						"Name": "/api",
						"State": {"Running": true},
						"NetworkSettings": {"Ports": {"3000/tcp": [{"HostPort": "3000"}]}}
					}]`), nil
				}
				return nil, errors.New("unexpected command")
			},
		},
	}

	containers, err := scanner.ListRunning(context.Background())
	if err != nil {
		t.Fatalf("ListRunning returned error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("got %d containers", len(containers))
	}
	if containers[0].Name != "api" || containers[1].Name != "web" {
		t.Fatalf("got order %#v", containers)
	}
}
