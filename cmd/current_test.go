package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestCurrent_PrintsLatest(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.01.0", "2025.03.2", "2025.03.1"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runCurrent(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "2025.03.2") {
		t.Errorf("expected latest tag 2025.03.2 in output, got: %q", out)
	}
}

func TestCurrent_NoTags_Error(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	err := runCurrent(nil)
	if err == nil {
		t.Fatal("expected error when no calver tags, got nil")
	}
	if !strings.Contains(err.Error(), "no CalVer tags") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCurrent_OnlyNonCalverTags_Error(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"v1.0.0", "release-2025", "latest"}, nil
		},
	}
	defer withMock(m)()

	if err := runCurrent(nil); err == nil {
		t.Error("expected error when no calver tags found among mixed tags")
	}
}

func TestCurrent_TagsError(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) { return nil, errSentinel },
	}
	defer withMock(m)()

	if err := runCurrent(nil); err == nil {
		t.Error("expected error when Tags() fails")
	}
}

// TestCurrent_HelpFlag_ViaExecute verifies --help exits cleanly (nil error)
// when dispatched through Execute(), which converts flag.ErrHelp → nil.
func TestCurrent_HelpFlag_ViaExecute(t *testing.T) {
	defer withMock(&mockOps{})()
	orig := os.Args
	os.Args = []string{"gh-calver", "current", "--help"}
	defer func() { os.Args = orig }()

	if err := Execute(); err != nil {
		t.Errorf("Execute 'current --help' should return nil, got: %v", err)
	}
}
