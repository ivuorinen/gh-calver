package cmd

import (
	"fmt"
	"os"

	"github.com/ivuorinen/gh-calver/internal/calver"
	"github.com/ivuorinen/gh-calver/internal/git"
)

const releaseHelp = `gh calver release — Create the next CalVer tag and publish a GitHub release

Usage:
  gh calver release [flags] [assets...]

Positional args:
  assets   One or more file paths to attach to the release as downloadable assets

Flags:
  --prefix      string   Optional tag prefix (e.g. "v" produces v2025.03.0)
  --title       string   Release title (defaults to the version string)
  --notes-file  string   Path to Markdown file for release notes
                         (uses GitHub auto-generated notes if omitted)
  --target      string   Branch or commit SHA to tag (defaults to HEAD)
  --message     string   Annotated tag message (lightweight tag if omitted)
  --draft                Save as draft — do not publish immediately
  --prerelease           Mark as a pre-release
  --dry-run              Print actions without executing them
  --help                 Show this help

Examples:
  gh calver release
  gh calver release --prefix v
  gh calver release --notes-file CHANGELOG.md
  gh calver release --draft --prerelease --prefix v
  gh calver release --title "March 2025" dist/app-linux.tar.gz dist/app-darwin.tar.gz
  gh calver release --dry-run`

func runRelease(args []string) error {
	fs := newFlagSet("release", releaseHelp)
	prefix := fs.String("prefix", "", `optional tag prefix (e.g. "v")`)
	title := fs.String("title", "", "release title (defaults to the version tag)")
	notesFile := fs.String("notes-file", "", "path to Markdown release notes file")
	target := fs.String("target", "", "branch or SHA to tag (default: HEAD)")
	message := fs.String("message", "", "annotated tag message")
	draft := fs.Bool("draft", false, "save as draft")
	prerelease := fs.Bool("prerelease", false, "mark as pre-release")
	dryRun := fs.Bool("dry-run", false, "print actions without executing them")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Remaining positional args are asset file paths.
	assets := fs.Args()

	// Validate notes file exists before doing any git work.
	if *notesFile != "" {
		if _, err := os.Stat(*notesFile); err != nil {
			return fmt.Errorf("notes file not found: %w", err)
		}
	}

	// Validate asset paths exist.
	for _, a := range assets {
		if _, err := os.Stat(a); err != nil {
			return fmt.Errorf("asset file not found: %w", err)
		}
	}

	tags, err := ops.Tags()
	if err != nil {
		return err
	}

	next := calver.Next(tags, nowFn(), *prefix)
	tag := next.String()

	if ops.TagExists(tag) {
		return fmt.Errorf("tag %q already exists — nothing to do", tag)
	}

	releaseTitle := *title
	if releaseTitle == "" {
		releaseTitle = tag
	}

	if *dryRun {
		printReleaseDryRun(tag, releaseTitle, *notesFile, *target, *draft, *prerelease, assets)
		return nil
	}

	// Step 1: create local tag.
	if err := ops.CreateTag(tag, *message); err != nil {
		return err
	}
	fmt.Printf("Created tag: %s\n", tag)

	// Step 2: push tag (clean up local on failure).
	if err := ops.PushTag(tag); err != nil {
		_ = ops.DeleteTag(tag)
		return fmt.Errorf("push failed (local tag removed): %w", err)
	}
	fmt.Printf("Pushed tag: %s\n", tag)

	// Step 3: create GitHub release.
	ropts := git.ReleaseOptions{
		Title:      releaseTitle,
		NotesFile:  *notesFile,
		Target:     *target,
		Draft:      *draft,
		Prerelease: *prerelease,
		Assets:     assets,
	}
	if err := ops.CreateRelease(tag, ropts); err != nil {
		return err
	}

	status := ""
	if *draft {
		status += " [draft]"
	}
	if *prerelease {
		status += " [pre-release]"
	}
	fmt.Printf("✓ Release %s published%s\n", tag, status)
	return nil
}

func printReleaseDryRun(
	tag, title, notesFile, target string,
	draft, prerelease bool,
	assets []string,
) {
	fmt.Println("[dry-run] The following actions would be taken:")
	fmt.Printf("  1. git tag %q\n", tag)
	fmt.Printf("  2. git push origin %q\n", tag)
	fmt.Printf("  3. gh release create %q\n", tag)
	fmt.Printf("     title:      %s\n", title)
	if notesFile != "" {
		fmt.Printf("     notes:      from file %q\n", notesFile)
	} else {
		fmt.Printf("     notes:      GitHub auto-generated\n")
	}
	if target != "" {
		fmt.Printf("     target:     %s\n", target)
	}
	if draft {
		fmt.Printf("     draft:      yes\n")
	}
	if prerelease {
		fmt.Printf("     prerelease: yes\n")
	}
	if len(assets) > 0 {
		fmt.Printf("     assets:     %v\n", assets)
	}
}
