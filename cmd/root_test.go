package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExecute_NoArgs_PrintsUsage(t *testing.T) {
	defer withMock(&mockOps{})()
	orig := os.Args
	os.Args = []string{"gh-calver"}
	defer func() { os.Args = orig }()

	out := captureStdout(func() {
		if err := Execute(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "gh-calver") {
		t.Errorf("expected usage in output, got: %q", out)
	}
}

func TestExecute_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h", "help"} {
		t.Run(flag, func(t *testing.T) {
			defer withMock(&mockOps{})()
			orig := os.Args
			os.Args = []string{"gh-calver", flag}
			defer func() { os.Args = orig }()

			out := captureStdout(func() {
				if err := Execute(); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, "Commands:") {
				t.Errorf("expected Commands: in help output, got: %q", out)
			}
		})
	}
}

func TestExecute_Version(t *testing.T) {
	for _, v := range []string{"version", "--version"} {
		t.Run(v, func(t *testing.T) {
			defer withMock(&mockOps{})()
			orig := os.Args
			os.Args = []string{"gh-calver", v}
			defer func() { os.Args = orig }()

			out := captureStdout(func() {
				if err := Execute(); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			if !strings.Contains(out, "gh-calver") {
				t.Errorf("expected version output, got: %q", out)
			}
		})
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	defer withMock(&mockOps{})()
	orig := os.Args
	os.Args = []string{"gh-calver", "foobar"}
	defer func() { os.Args = orig }()

	err := Execute()
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
	if !strings.Contains(err.Error(), "foobar") {
		t.Errorf("expected command name in error, got: %v", err)
	}
}

func TestExecute_HelpFlagExitsClean(t *testing.T) {
	// Passing --help to a subcommand should return nil (not flag.ErrHelp).
	defer withMock(&mockOps{})()
	orig := os.Args
	os.Args = []string{"gh-calver", "next", "--help"}
	defer func() { os.Args = orig }()

	if err := Execute(); err != nil {
		t.Errorf("--help should return nil, got: %v", err)
	}
}

func TestExecute_LsAlias(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()
	orig := os.Args
	os.Args = []string{"gh-calver", "ls"}
	defer func() { os.Args = orig }()

	out := captureStdout(func() {
		if err := Execute(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No CalVer tags") {
		t.Errorf("expected list output via ls alias, got: %q", out)
	}
}

func TestExecute_DispatchesCurrent(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) { return []string{"2025.03.0"}, nil },
	}
	defer withMock(m)()
	orig := os.Args
	os.Args = []string{"gh-calver", "current"}
	defer func() { os.Args = orig }()

	out := captureStdout(func() {
		if err := Execute(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "2025.03.0") {
		t.Errorf("expected calver output, got: %q", out)
	}
}

func TestExecute_DispatchesBump(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()
	orig := os.Args
	os.Args = []string{"gh-calver", "bump", "--dry-run"}
	defer func() { os.Args = orig }()

	out := captureStdout(func() {
		if err := Execute(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected dry-run output, got: %q", out)
	}
}

func TestExecute_DispatchesRelease(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()
	orig := os.Args
	os.Args = []string{"gh-calver", "release", "--dry-run"}
	defer func() { os.Args = orig }()

	out := captureStdout(func() {
		if err := Execute(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected dry-run output, got: %q", out)
	}
}
