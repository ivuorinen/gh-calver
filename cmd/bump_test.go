package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestBump_DryRun(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runBump([]string{"--dry-run"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected dry-run output, got: %q", out)
	}
	if !strings.Contains(out, "2025.03.0") {
		t.Errorf("expected tag in dry-run output, got: %q", out)
	}
	if len(m.CreatedTags) != 0 {
		t.Errorf("dry-run should not create tags, got: %v", m.CreatedTags)
	}
}

func TestBump_DryRunWithPrefix(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runBump([]string{"--dry-run", "--prefix", "v"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "v2025.03.0") {
		t.Errorf("expected v-prefixed tag in dry-run output, got: %q", out)
	}
}

func TestBump_Success(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runBump(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.CreatedTags) != 1 || m.CreatedTags[0] != "2025.03.0" {
		t.Errorf("expected tag 2025.03.0 created, got: %v", m.CreatedTags)
	}
	if len(m.PushedTags) != 1 || m.PushedTags[0] != "2025.03.0" {
		t.Errorf("expected tag 2025.03.0 pushed, got: %v", m.PushedTags)
	}
	if !strings.Contains(out, "Created tag") {
		t.Errorf("expected 'Created tag' in output, got: %q", out)
	}
	if !strings.Contains(out, "Pushed tag") {
		t.Errorf("expected 'Pushed tag' in output, got: %q", out)
	}
}

func TestBump_TagAlreadyExists(t *testing.T) {
	m := &mockOps{
		OnTags:      func() ([]string, error) { return []string{"2025.03.0"}, nil },
		OnTagExists: func(tag string) bool { return tag == "2025.03.1" },
	}
	// The next tag would be 2025.03.1 (collision with existing 2025.03.0)
	// but we mark 2025.03.1 as already existing too.
	m.OnTags = func() ([]string, error) { return []string{"2025.03.0"}, nil }
	m.OnTagExists = func(tag string) bool { return true } // always exists
	defer withMock(m)()

	err := runBump(nil)
	if err == nil {
		t.Error("expected error when tag already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBump_CreateTagFails(t *testing.T) {
	m := &mockOps{
		OnTags:      func() ([]string, error) { return nil, nil },
		OnCreateTag: func(tag, msg string) error { return errors.New("tag create failed") },
	}
	defer withMock(m)()

	if err := runBump(nil); err == nil {
		t.Error("expected error when CreateTag fails")
	}
}

func TestBump_PushFailsCleansUpLocalTag(t *testing.T) {
	m := &mockOps{
		OnTags:    func() ([]string, error) { return nil, nil },
		OnPushTag: func(tag string) error { return errors.New("push failed") },
	}
	defer withMock(m)()

	err := runBump(nil)
	if err == nil {
		t.Error("expected error when push fails")
	}
	if !strings.Contains(err.Error(), "push failed") {
		t.Errorf("unexpected error: %v", err)
	}
	// Tag should have been deleted (cleanup on push failure)
	if len(m.DeletedTags) != 1 {
		t.Errorf("expected 1 tag deleted on push failure, got: %v", m.DeletedTags)
	}
}

func TestBump_TagsError(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, errSentinel }}
	defer withMock(m)()

	if err := runBump(nil); err == nil {
		t.Error("expected error when Tags() fails")
	}
}

func TestBump_InvalidFlag(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	err := runBump([]string{"--bogus"})
	if err == nil {
		t.Error("expected parse error for unknown flag")
	}
}

func TestBump_MessagePassedToCreateTag(t *testing.T) {
	var gotMsg string
	m := &mockOps{
		OnTags:      func() ([]string, error) { return nil, nil },
		OnCreateTag: func(tag, msg string) error { gotMsg = msg; return nil },
	}
	defer withMock(m)()

	captureStdout(func() {
		if err := runBump([]string{"--message", "annotated message"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if gotMsg != "annotated message" {
		t.Errorf("expected message 'annotated message', got %q", gotMsg)
	}
}
