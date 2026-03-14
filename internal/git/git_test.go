package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// initRepo creates a temporary git repository and returns its path and a
// cleanup function. The repo has at least one commit so tags can be created.
func initRepo(t *testing.T) (dir string) {
	t.Helper()
	dir = t.TempDir()

	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", c, err, out)
		}
	}
	return dir
}

// inDir changes the working directory to dir for the duration of the test.
func inDir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// gitTag creates a raw git tag directly without going through our code.
func gitTag(t *testing.T, dir, tag string) {
	t.Helper()
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag %s: %v\n%s", tag, err, out)
	}
}

// fakeBinDir creates a temporary directory with a fake script for `name` that
// exits 0 and prints output. It prepends the dir to PATH and returns cleanup.
func fakeBinDir(t *testing.T, name, script string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	orig := os.Getenv("PATH")
	os.Setenv("PATH", dir+":"+orig)
	t.Cleanup(func() { os.Setenv("PATH", orig) })
}

// ── Tags ─────────────────────────────────────────────────────────────────────

func TestTags_Empty(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	tags, err := Tags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %v", tags)
	}
}

func TestTags_WithTags(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	for _, tag := range []string{"2025.03.0", "v2025.02.0", "not-calver"} {
		gitTag(t, dir, tag)
	}

	tags, err := Tags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d: %v", len(tags), tags)
	}
}

func TestTags_NotARepo(t *testing.T) {
	inDir(t, t.TempDir()) // empty dir, not a git repo
	_, err := Tags()
	if err == nil {
		t.Error("expected error for non-repo directory, got nil")
	}
}

// ── TagExists ─────────────────────────────────────────────────────────────────

func TestTagExists_Present(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	if !TagExists("2025.03.0") {
		t.Error("expected TagExists to return true for existing tag")
	}
}

func TestTagExists_Absent(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	if TagExists("2025.03.0") {
		t.Error("expected TagExists to return false for missing tag")
	}
}

// ── CreateTag ─────────────────────────────────────────────────────────────────

func TestCreateTag_Lightweight(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	if err := CreateTag("2025.03.0", ""); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if !TagExists("2025.03.0") {
		t.Error("tag not found after CreateTag")
	}
}

func TestCreateTag_Annotated(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	if err := CreateTag("2025.03.0", "Release 2025.03.0"); err != nil {
		t.Fatalf("CreateTag annotated: %v", err)
	}
	if !TagExists("2025.03.0") {
		t.Error("annotated tag not found after CreateTag")
	}
}

func TestCreateTag_DuplicateErrors(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	err := CreateTag("2025.03.0", "")
	if err == nil {
		t.Error("expected error creating duplicate tag, got nil")
	}
}

// ── DeleteTag ─────────────────────────────────────────────────────────────────

func TestDeleteTag(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	if err := DeleteTag("2025.03.0"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	if TagExists("2025.03.0") {
		t.Error("tag still exists after DeleteTag")
	}
}

func TestDeleteTag_Missing(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	err := DeleteTag("does-not-exist")
	if err == nil {
		t.Error("expected error deleting non-existent tag, got nil")
	}
}

// ── CurrentBranch ─────────────────────────────────────────────────────────────

func TestCurrentBranch(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	branch, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

// ── PushTag ────────────────────────────────────────────────────────────────────

func TestPushTag_NoRemote(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	err := PushTag("2025.03.0")
	if err == nil {
		t.Error("expected error pushing tag with no remote, got nil")
	}
	if !strings.Contains(err.Error(), "pushing tag") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPushTag_ToLocalRemote(t *testing.T) {
	// Use a second bare repo as the "remote" so the push succeeds.
	remote := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", remote)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init bare: %v\n%s", err, out)
	}

	dir := initRepo(t)
	inDir(t, dir)

	// Add the bare repo as origin.
	addRemote := exec.Command("git", "remote", "add", "origin", remote)
	addRemote.Dir = dir
	if out, err := addRemote.CombinedOutput(); err != nil {
		t.Fatalf("add remote: %v\n%s", err, out)
	}

	// Push main branch first so the tag has something to reference.
	pushMain := exec.Command("git", "push", "origin", "main")
	pushMain.Dir = dir
	if out, err := pushMain.CombinedOutput(); err != nil {
		t.Fatalf("push main: %v\n%s", err, out)
	}

	gitTag(t, dir, "2025.03.0")

	if err := PushTag("2025.03.0"); err != nil {
		t.Fatalf("PushTag: %v", err)
	}
}

// ── CreateRelease (fake gh) ────────────────────────────────────────────────────

func TestCreateRelease_DefaultNotes(t *testing.T) {
	var gotArgs []string
	fakeBinDir(t, "gh", fmt.Sprintf(
		// Write args to a temp file so we can inspect them.
		`echo "$@" > %s`,
		filepath.Join(t.TempDir(), "gh-args.txt"),
	))
	// Use a capturing fake gh instead.
	captureDir := t.TempDir()
	argsFile := filepath.Join(captureDir, "gh-args.txt")
	script := fmt.Sprintf(`printf "%%s\n" "$@" > %s`, argsFile)
	fakeBinDir(t, "gh", script)

	opts := ReleaseOptions{Title: "Test release"}
	if err := CreateRelease("2025.03.0", opts); err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	raw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("reading args file: %v", err)
	}
	gotArgs = strings.Split(strings.TrimSpace(string(raw)), "\n")

	wantSubset := []string{"release", "create", "2025.03.0", "--generate-notes"}
	for _, w := range wantSubset {
		found := slices.Contains(gotArgs, w)
		if !found {
			t.Errorf("expected arg %q in %v", w, gotArgs)
		}
	}
}

func TestCreateRelease_WithNotesFile(t *testing.T) {
	captureDir := t.TempDir()
	argsFile := filepath.Join(captureDir, "gh-args.txt")
	fakeBinDir(t, "gh", fmt.Sprintf(`printf "%%s\n" "$@" > %s`, argsFile))

	opts := ReleaseOptions{
		Title:      "Test",
		NotesFile:  "/tmp/notes.md",
		Draft:      true,
		Prerelease: true,
		Target:     "main",
		Assets:     []string{"app.tar.gz"},
	}
	if err := CreateRelease("2025.03.0", opts); err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	raw, _ := os.ReadFile(argsFile)
	out := string(raw)

	for _, want := range []string{"--notes-file", "--draft", "--prerelease", "--target", "app.tar.gz"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in gh args:\n%s", want, out)
		}
	}
	if strings.Contains(out, "--generate-notes") {
		t.Errorf("should NOT have --generate-notes when notes-file is set")
	}
}

func TestCreateRelease_GhFails(t *testing.T) {
	fakeBinDir(t, "gh", "exit 1")
	err := CreateRelease("2025.03.0", ReleaseOptions{})
	if err == nil {
		t.Error("expected error when gh exits 1, got nil")
	}
	if !strings.Contains(err.Error(), "creating GitHub release") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ── Client interface ──────────────────────────────────────────────────────────

func TestNewClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that realClient satisfies Client.
	var _ Client = NewClient()
}

// ── realClient delegation ─────────────────────────────────────────────────────

func TestRealClient_Tags(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")
	gitTag(t, dir, "2025.03.1")

	c := NewClient()
	tags, err := c.Tags()
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d: %v", len(tags), tags)
	}
}

func TestRealClient_TagExists(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	c := NewClient()
	if !c.TagExists("2025.03.0") {
		t.Error("expected TagExists to return true")
	}
	if c.TagExists("nonexistent") {
		t.Error("expected TagExists to return false")
	}
}

func TestRealClient_CreateTag(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)

	c := NewClient()
	if err := c.CreateTag("2025.03.0", ""); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if !TagExists("2025.03.0") {
		t.Error("tag not found after CreateTag via client")
	}
}

func TestRealClient_DeleteTag(t *testing.T) {
	dir := initRepo(t)
	inDir(t, dir)
	gitTag(t, dir, "2025.03.0")

	c := NewClient()
	if err := c.DeleteTag("2025.03.0"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	if TagExists("2025.03.0") {
		t.Error("tag still exists after DeleteTag via client")
	}
}

func TestRealClient_PushTag(t *testing.T) {
	remote := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", remote)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init bare: %v\n%s", err, out)
	}

	dir := initRepo(t)
	inDir(t, dir)

	addRemote := exec.Command("git", "remote", "add", "origin", remote)
	addRemote.Dir = dir
	if out, err := addRemote.CombinedOutput(); err != nil {
		t.Fatalf("add remote: %v\n%s", err, out)
	}

	pushMain := exec.Command("git", "push", "origin", "main")
	pushMain.Dir = dir
	if out, err := pushMain.CombinedOutput(); err != nil {
		t.Fatalf("push main: %v\n%s", err, out)
	}

	gitTag(t, dir, "2025.03.0")

	c := NewClient()
	if err := c.PushTag("2025.03.0"); err != nil {
		t.Fatalf("PushTag via client: %v", err)
	}
}

func TestRealClient_CreateRelease(t *testing.T) {
	// Without a real gh CLI configured, this will error — but it covers the code path.
	c := NewClient()
	err := c.CreateRelease("2025.03.0", ReleaseOptions{Title: "test"})
	if err == nil {
		t.Log("CreateRelease unexpectedly succeeded (gh CLI available?)")
	}
}

func TestCurrentBranch_NotARepo(t *testing.T) {
	inDir(t, t.TempDir())
	_, err := CurrentBranch()
	if err == nil {
		t.Error("expected error for CurrentBranch outside a git repo")
	}
	if !strings.Contains(err.Error(), "getting current branch") {
		t.Errorf("unexpected error message: %v", err)
	}
}
