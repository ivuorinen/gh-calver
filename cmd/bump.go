package cmd

import (
	"fmt"

	"github.com/ivuorinen/gh-calver/internal/calver"
)

const bumpHelp = `gh calver bump — Create and push the next CalVer tag (no GitHub release)

Usage:
  gh calver bump [flags]

Flags:
  --prefix   string   Optional tag prefix (e.g. "v")
  --message  string   Annotated tag message (lightweight tag if omitted)
  --dry-run           Print actions without executing them
  --help              Show this help

Examples:
  gh calver bump
  gh calver bump --prefix v --message "March 2025 release"
  gh calver bump --dry-run`

func runBump(args []string) error {
	fs := newFlagSet("bump", bumpHelp)
	prefix := fs.String("prefix", "", `optional tag prefix (e.g. "v")`)
	message := fs.String("message", "", "annotated tag message")
	dryRun := fs.Bool("dry-run", false, "print actions without executing them")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tags, err := ops.Tags()
	if err != nil {
		return err
	}

	next := calver.Next(tags, nowFn(), *prefix)
	tag := next.String()

	if ops.TagExists(tag) {
		return fmt.Errorf("tag %q already exists", tag)
	}

	if *dryRun {
		fmt.Printf("[dry-run] would create tag: %s\n", tag)
		fmt.Printf("[dry-run] would push tag to origin\n")
		return nil
	}

	if err := ops.CreateTag(tag, *message); err != nil {
		return err
	}
	fmt.Printf("Created tag: %s\n", tag)

	if err := ops.PushTag(tag); err != nil {
		_ = ops.DeleteTag(tag)
		return fmt.Errorf("push failed (local tag removed): %w", err)
	}
	fmt.Printf("Pushed tag %s to origin\n", tag)
	return nil
}
