package cmd

import (
	"strings"
	"testing"
)

func TestNext_NoExistingTags(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runNext(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// testTime is 2025-03-15, so first release of month is 2025.03.0
	want := "2025.03.0"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q in output, got: %q", want, out)
	}
}

func TestNext_CollisionIncrement(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.03.0", "2025.03.1"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runNext(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	want := "2025.03.2"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q in output, got: %q", want, out)
	}
}

func TestNext_WithPrefix(t *testing.T) {
	m := &mockOps{
		OnTags: func() ([]string, error) { return []string{"v2025.03.0"}, nil },
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runNext([]string{"--prefix", "v"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	want := "v2025.03.1"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q in output, got: %q", want, out)
	}
}

func TestNext_TagsError(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, errSentinel }}
	defer withMock(m)()

	if err := runNext(nil); err == nil {
		t.Error("expected error when Tags() fails")
	}
}

func TestNext_PreviousMonthTagsReset(t *testing.T) {
	// Tags from February should not affect March micro counter.
	m := &mockOps{
		OnTags: func() ([]string, error) {
			return []string{"2025.02.0", "2025.02.1", "2025.02.5"}, nil
		},
	}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runNext(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	want := "2025.03.0"
	if !strings.Contains(out, want) {
		t.Errorf("expected %q (reset to 0 for new month), got: %q", want, out)
	}
}
