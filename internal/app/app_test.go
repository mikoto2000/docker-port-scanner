package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type stubScanner struct {
	listFn func(context.Context) ([]Container, error)
	getFn  func(context.Context, string) (Container, error)
}

func (s stubScanner) ListRunning(ctx context.Context) ([]Container, error) {
	return s.listFn(ctx)
}

func (s stubScanner) GetContainer(ctx context.Context, reference string) (Container, error) {
	return s.getFn(ctx, reference)
}

func TestBuildURL(t *testing.T) {
	got := BuildURL("https", "127.0.0.1", 8443)
	want := "https://127.0.0.1:8443"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderOutputSortsPortsAndSeparatesContainers(t *testing.T) {
	containers := []Container{
		{Name: "web", Ports: []int{8443, 8080}},
		{Name: "api", Ports: []int{3000}},
	}

	got, err := RenderOutput(containers, "http", "localhost")
	if err != nil {
		t.Fatalf("RenderOutput returned error: %v", err)
	}

	want := strings.Join([]string{
		"web",
		"http://localhost:8080",
		"http://localhost:8443",
		"",
		"api",
		"http://localhost:3000",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRunUsesSpecifiedTargetsInOrder(t *testing.T) {
	var out bytes.Buffer

	scanner := stubScanner{
		listFn: func(context.Context) ([]Container, error) {
			t.Fatal("ListRunning should not be called")
			return nil, nil
		},
		getFn: func(_ context.Context, reference string) (Container, error) {
			switch reference {
			case "web":
				return Container{Name: "frontend", Ports: []int{8080}}, nil
			case "api":
				return Container{Name: "backend", Ports: []int{3000}}, nil
			default:
				return Container{}, errors.New("unexpected target")
			}
		},
	}

	err := Run(context.Background(), []string{"--host", "127.0.0.1", "--scheme", "https", "web", "api"}, &out, &bytes.Buffer{}, scanner)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := strings.Join([]string{
		"frontend",
		"https://127.0.0.1:8080",
		"",
		"backend",
		"https://127.0.0.1:3000",
		"",
	}, "\n")
	if out.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRunListsRunningContainersAndSkipsWithoutPorts(t *testing.T) {
	var out bytes.Buffer

	scanner := stubScanner{
		listFn: func(context.Context) ([]Container, error) {
			return []Container{
				{Name: "worker"},
				{Name: "web", Ports: []int{8080}},
			}, nil
		},
		getFn: func(context.Context, string) (Container, error) {
			t.Fatal("GetContainer should not be called")
			return Container{}, nil
		},
	}

	err := Run(context.Background(), nil, &out, &bytes.Buffer{}, scanner)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := strings.Join([]string{
		"web",
		"http://localhost:8080",
		"",
	}, "\n")
	if out.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRunReturnsErrorWhenSpecifiedContainerHasNoPorts(t *testing.T) {
	scanner := stubScanner{
		listFn: func(context.Context) ([]Container, error) {
			return nil, nil
		},
		getFn: func(context.Context, string) (Container, error) {
			return Container{Name: "db"}, nil
		},
	}

	err := Run(context.Background(), []string{"db"}, &bytes.Buffer{}, &bytes.Buffer{}, scanner)
	if err == nil || err.Error() != `container "db" has no published TCP ports` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsErrorWhenNoContainersCanBeRendered(t *testing.T) {
	scanner := stubScanner{
		listFn: func(context.Context) ([]Container, error) {
			return []Container{{Name: "worker"}}, nil
		},
		getFn: func(context.Context, string) (Container, error) {
			return Container{}, nil
		},
	}

	err := Run(context.Background(), nil, &bytes.Buffer{}, &bytes.Buffer{}, scanner)
	if err == nil || err.Error() != "no running containers expose published TCP ports" {
		t.Fatalf("unexpected error: %v", err)
	}
}
