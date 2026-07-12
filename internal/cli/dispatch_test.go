package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/igor-kupczynski/gw-code/internal/app"
)

func TestRootHelpOmitsCompletion(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	root := buildRoot(&app.App{})
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	if strings.Contains(strings.ToLower(help), "completion") {
		t.Fatalf("root help lists completion:\n%s", help)
	}
}

func TestRootHelpListsModeFlags(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	root := buildRoot(&app.App{})
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	for _, want := range []string{"--refresh", "--path"} {
		if !strings.Contains(help, want) {
			t.Fatalf("root help missing %q:\n%s", want, help)
		}
	}
	for _, forbidden := range []string{"--editor", "--window", "--wait", "--provider"} {
		if strings.Contains(help, forbidden) {
			t.Fatalf("root help lists unsupported flag %q:\n%s", forbidden, help)
		}
	}
}

func TestRefreshAndPathMutuallyExclusive(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	root := buildRoot(&app.App{})
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--refresh", "--path", "demo"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "refresh") || !strings.Contains(err.Error(), "path") {
		t.Fatalf("err = %v", err)
	}
}
