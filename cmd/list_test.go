package cmd

import (
	"strings"
	"testing"
)

func TestList_PrintsAllSorted(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.01.0", "2025.03.1", "2025.03.0", "2024.12.5"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
	}
	// Newest first.
	if lines[0] != "2025.03.1" {
		t.Errorf("expected first line 2025.03.1, got %q", lines[0])
	}
	if lines[len(lines)-1] != "2024.12.5" {
		t.Errorf("expected last line 2024.12.5, got %q", lines[len(lines)-1])
	}
}

func TestList_NoCalverTags(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) { return []string{"v1.0.0", "release"}, nil },
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No CalVer tags") {
		t.Errorf("expected 'No CalVer tags' message, got: %q", out)
	}
}

func TestList_Empty(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No CalVer tags") {
		t.Errorf("expected empty-state message, got: %q", out)
	}
}

func TestList_WithLimit(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.03.0", "2025.02.0", "2025.01.0", "2024.12.0"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList([]string{"--limit", "2"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with --limit 2, got %d: %v", len(lines), lines)
	}
}

func TestList_LimitZeroMeansAll(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.03.0", "2025.02.0", "2025.01.0"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList([]string{"--limit", "0"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected all 3 lines with --limit 0, got %d: %v", len(lines), lines)
	}
}

func TestList_LimitExceedsTotal(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.03.0", "2025.02.0"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runList([]string{"--limit", "100"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("limit > total should show all; expected 2, got %d: %v", len(lines), lines)
	}
}

func TestList_TagsError(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, errSentinel }}
	defer withMock(m)()

	if err := runList(nil); err == nil {
		t.Error("expected error when Tags() fails")
	}
}

func TestList_InvalidLimitFlag(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()
	// Passing a non-integer to --limit should return a parse error, not panic.
	err := runList([]string{"--limit", "notanint"})
	if err == nil {
		t.Error("expected parse error for non-integer --limit")
	}
}
