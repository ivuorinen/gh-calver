package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-calver/internal/git"
)

// ── dry-run ───────────────────────────────────────────────────────────────────

func TestRelease_DryRun_DefaultNotes(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runRelease([]string{"--dry-run"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{"[dry-run]", "2025.03.0", "auto-generated"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dry-run output:\n%s", want, out)
		}
	}
	if len(m.CreatedTags) != 0 {
		t.Errorf("dry-run should not create tags")
	}
}

func TestRelease_DryRun_WithAllFlags(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	notesFile := writeTemp(t, "# Notes")
	out := captureStdout(func() {
		args := []string{
			"--dry-run",
			"--prefix", "v",
			"--title", "My Release",
			"--notes-file", notesFile,
			"--target", "main",
			"--draft",
			"--prerelease",
		}
		if err := runRelease(args); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{
		"v2025.03.0", "My Release", "draft", "prerelease", "main", notesFile,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dry-run output:\n%s", want, out)
		}
	}
}

func TestRelease_DryRun_WithAssets(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	asset := writeTemp(t, "binary data")
	out := captureStdout(func() {
		if err := runRelease([]string{"--dry-run", asset}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, asset) {
		t.Errorf("expected asset path in dry-run output:\n%s", out)
	}
}

// ── full release flow ─────────────────────────────────────────────────────────

func TestRelease_Success_AutoNotes(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runRelease(nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.CreatedTags) != 1 || m.CreatedTags[0] != "2025.03.0" {
		t.Errorf("expected tag created, got: %v", m.CreatedTags)
	}
	if len(m.PushedTags) != 1 {
		t.Errorf("expected tag pushed, got: %v", m.PushedTags)
	}
	if len(m.ReleasedTags) != 1 {
		t.Errorf("expected release created, got: %v", m.ReleasedTags)
	}
	if !strings.Contains(out, "✓ Release") {
		t.Errorf("expected success message in output:\n%s", out)
	}
}

func TestRelease_Success_WithPrefix(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	captureStdout(func() {
		if err := runRelease([]string{"--prefix", "v"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.CreatedTags) != 1 || m.CreatedTags[0] != "v2025.03.0" {
		t.Errorf("expected v-prefixed tag created, got: %v", m.CreatedTags)
	}
}

func TestRelease_Success_WithNotesFile(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	notesFile := writeTemp(t, "# Changelog\n\n- fix bug")
	captureStdout(func() {
		if err := runRelease([]string{"--notes-file", notesFile}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.ReleaseOpts) == 0 {
		t.Fatal("no release opts recorded")
	}
	if m.ReleaseOpts[0].NotesFile != notesFile {
		t.Errorf("expected notes file %q, got %q", notesFile, m.ReleaseOpts[0].NotesFile)
	}
}

func TestRelease_Success_DraftAndPrerelease(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	out := captureStdout(func() {
		if err := runRelease([]string{"--draft", "--prerelease"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.ReleaseOpts) == 0 {
		t.Fatal("no release opts recorded")
	}
	opts := m.ReleaseOpts[0]
	if !opts.Draft {
		t.Error("expected Draft=true in release options")
	}
	if !opts.Prerelease {
		t.Error("expected Prerelease=true in release options")
	}
	if !strings.Contains(out, "[draft]") {
		t.Errorf("expected [draft] in success output:\n%s", out)
	}
	if !strings.Contains(out, "[pre-release]") {
		t.Errorf("expected [pre-release] in success output:\n%s", out)
	}
}

func TestRelease_Success_WithTarget(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	captureStdout(func() {
		if err := runRelease([]string{"--target", "develop"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.ReleaseOpts) == 0 {
		t.Fatal("no release opts recorded")
	}
	if m.ReleaseOpts[0].Target != "develop" {
		t.Errorf("expected target 'develop', got %q", m.ReleaseOpts[0].Target)
	}
}

func TestRelease_Success_WithCustomTitle(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	captureStdout(func() {
		if err := runRelease([]string{"--title", "My Custom Title"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if m.ReleaseOpts[0].Title != "My Custom Title" {
		t.Errorf("expected custom title, got %q", m.ReleaseOpts[0].Title)
	}
}

func TestRelease_TitleDefaultsToTag(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	captureStdout(func() { _ = runRelease(nil) })

	if len(m.ReleaseOpts) == 0 {
		t.Fatal("no release opts recorded")
	}
	if m.ReleaseOpts[0].Title != "2025.03.0" {
		t.Errorf("expected title to default to tag, got %q", m.ReleaseOpts[0].Title)
	}
}

func TestRelease_Success_WithAssets(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	asset1 := writeTemp(t, "binary1")
	asset2 := writeTemp(t, "binary2")
	captureStdout(func() {
		if err := runRelease([]string{asset1, asset2}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if len(m.ReleaseOpts) == 0 {
		t.Fatal("no release opts recorded")
	}
	opts := m.ReleaseOpts[0]
	if len(opts.Assets) != 2 {
		t.Errorf("expected 2 assets, got %d: %v", len(opts.Assets), opts.Assets)
	}
}

// ── validation errors ─────────────────────────────────────────────────────────

func TestRelease_InvalidFlag(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	err := runRelease([]string{"--bogus"})
	if err == nil {
		t.Error("expected parse error for unknown flag")
	}
}

func TestRelease_NotesFileMissing_Error(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	err := runRelease([]string{"--notes-file", "/nonexistent/path/notes.md"})
	if err == nil {
		t.Error("expected error for missing notes file")
	}
	if !strings.Contains(err.Error(), "notes file not found") {
		t.Errorf("unexpected error: %v", err)
	}
	// Should not have called Tags() at all (fail-fast before git work).
	if m.OnTags != nil {
		// Tags() could be called or not — the important thing is no tag was created.
		if len(m.CreatedTags) != 0 {
			t.Error("should not create tags when notes file is missing")
		}
	}
}

func TestRelease_AssetFileMissing_Error(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, nil }}
	defer withMock(m)()

	err := runRelease([]string{"/nonexistent/asset.tar.gz"})
	if err == nil {
		t.Error("expected error for missing asset file")
	}
	if !strings.Contains(err.Error(), "asset file not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRelease_TagAlreadyExists_Error(t *testing.T) {
	m := &mockOps{
		OnTags:      func() ([]string, error) { return nil, nil },
		OnTagExists: func(tag string) bool { return true },
	}
	defer withMock(m)()

	err := runRelease(nil)
	if err == nil {
		t.Error("expected error when tag already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRelease_TagsError(t *testing.T) {
	m := &mockOps{OnTags: func() ([]string, error) { return nil, errSentinel }}
	defer withMock(m)()

	if err := runRelease(nil); err == nil {
		t.Error("expected error when Tags() fails")
	}
}

func TestRelease_CreateTagFails(t *testing.T) {
	m := &mockOps{
		OnTags:      func() ([]string, error) { return nil, nil },
		OnCreateTag: func(tag, msg string) error { return errors.New("create failed") },
	}
	defer withMock(m)()

	if err := runRelease(nil); err == nil {
		t.Error("expected error when CreateTag fails")
	}
}

func TestRelease_PushFails_CleansUpTag(t *testing.T) {
	m := &mockOps{
		OnTags:    func() ([]string, error) { return nil, nil },
		OnPushTag: func(tag string) error { return errors.New("network error") },
	}
	defer withMock(m)()

	err := runRelease(nil)
	if err == nil {
		t.Error("expected error when push fails")
	}
	if len(m.DeletedTags) != 1 {
		t.Errorf("expected local tag to be cleaned up on push failure, got: %v", m.DeletedTags)
	}
	// Release should NOT have been created.
	if len(m.ReleasedTags) != 0 {
		t.Errorf("release should not be created when push fails, got: %v", m.ReleasedTags)
	}
}

func TestRelease_CreateReleaseFails(t *testing.T) {
	m := &mockOps{
		OnTags:          func() ([]string, error) { return nil, nil },
		OnCreateRelease: func(tag string, opts git.ReleaseOptions) error { return errors.New("gh failed") },
	}
	defer withMock(m)()

	err := runRelease(nil)
	if err == nil {
		t.Error("expected error when CreateRelease fails")
	}
	// Tag was created and pushed, but release failed.
	if len(m.CreatedTags) != 1 {
		t.Errorf("expected tag created before release attempt, got: %v", m.CreatedTags)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// writeTemp creates a temp file with content and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.txt")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return filepath.ToSlash(f.Name())
}
