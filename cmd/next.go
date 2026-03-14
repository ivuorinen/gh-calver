package cmd

import (
	"fmt"

	"github.com/ivuorinen/gh-calver/internal/calver"
)

const nextHelp = `gh calver next — Preview the next CalVer tag without creating it

Usage:
  gh calver next [flags]

Flags:
  --prefix string   Optional tag prefix (e.g. "v" produces v2025.03.0)
  --help            Show this help

Examples:
  gh calver next
  gh calver next --prefix v`

func runNext(args []string) error {
	fs := newFlagSet("next", nextHelp)
	prefix := fs.String("prefix", "", `optional tag prefix (e.g. "v")`)
	if err := fs.Parse(args); err != nil {
		return err
	}

	tags, err := ops.Tags()
	if err != nil {
		return err
	}

	next := calver.Next(tags, nowFn(), *prefix)
	fmt.Println(next)
	return nil
}
