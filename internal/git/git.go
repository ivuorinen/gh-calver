// Package git wraps git and gh CLI operations needed by gh-calver.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Tags returns all tags in the current git repository.
func Tags() ([]string, error) {
	out, err := run("git", "tag", "-l")
	if err != nil {
		return nil, fmt.Errorf("listing git tags: %w", err)
	}
	raw := strings.TrimSpace(out)
	if raw == "" {
		return nil, nil
	}
	lines := strings.Split(raw, "\n")
	tags := make([]string, 0, len(lines))
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// CreateTag creates a local annotated git tag with an optional message.
// If message is empty, a lightweight tag is created instead.
func CreateTag(tag, message string) error {
	var args []string
	if message != "" {
		args = []string{"tag", "-a", tag, "-m", message}
	} else {
		args = []string{"tag", tag}
	}
	if _, err := run("git", args...); err != nil {
		return fmt.Errorf("creating tag %q: %w", tag, err)
	}
	return nil
}

// PushTag pushes a single tag to origin.
func PushTag(tag string) error {
	if _, err := run("git", "push", "origin", tag); err != nil {
		return fmt.Errorf("pushing tag %q: %w", tag, err)
	}
	return nil
}

// DeleteTag removes a local tag (useful for cleanup on failure).
func DeleteTag(tag string) error {
	_, err := run("git", "tag", "-d", tag)
	return err
}

// CurrentBranch returns the currently checked-out branch name.
func CurrentBranch() (string, error) {
	out, err := run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// ReleaseOptions configures how a GitHub release is created via gh.
type ReleaseOptions struct {
	Title      string
	NotesFile  string
	Target     string
	Draft      bool
	Prerelease bool
	Assets     []string // paths to upload as release assets
}

// CreateRelease creates a GitHub release for the given tag via `gh release create`.
func CreateRelease(tag string, opts ReleaseOptions) error {
	args := []string{"release", "create", tag}

	if opts.Title != "" {
		args = append(args, "--title", opts.Title)
	} else {
		args = append(args, "--title", tag)
	}

	if opts.NotesFile != "" {
		args = append(args, "--notes-file", opts.NotesFile)
	} else {
		args = append(args, "--generate-notes")
	}

	if opts.Draft {
		args = append(args, "--draft")
	}
	if opts.Prerelease {
		args = append(args, "--prerelease")
	}
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	// Append any asset file paths at the end (gh release create syntax)
	args = append(args, opts.Assets...)

	if _, err := runGH(args...); err != nil {
		return fmt.Errorf("creating GitHub release: %w", err)
	}
	return nil
}

// TagExists returns true if a tag with the given name exists locally.
func TagExists(tag string) bool {
	_, err := run("git", "rev-parse", "--verify", "refs/tags/"+tag)
	return err == nil
}

// run executes a command and returns combined stdout, or an error that
// includes stderr for better diagnostics.
func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", err
		}
		return "", fmt.Errorf("%w\n%s", err, msg)
	}
	return stdout.String(), nil
}

// runGH executes `gh` and returns stdout, including stderr in errors.
func runGH(args ...string) (string, error) {
	return run("gh", args...)
}
