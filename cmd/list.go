package cmd

import (
	"fmt"

	"github.com/ivuorinen/gh-calver/internal/calver"
)

const listHelp = `gh calver list — List all CalVer tags in this repository

Usage:
  gh calver list [flags]

Alias: ls

Flags:
  --limit  int   Maximum number of tags to show (0 = no limit)
  --help         Show this help

Examples:
  gh calver list
  gh calver list --limit 10`

func runList(args []string) error {
	fs := newFlagSet("list", listHelp)
	limit := fs.Int("limit", 0, "max number of tags to show (0 = all)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tags, err := ops.Tags()
	if err != nil {
		return err
	}

	versions := calver.All(tags)
	if len(versions) == 0 {
		fmt.Println("No CalVer tags found in this repository.")
		return nil
	}

	n := len(versions)
	if *limit > 0 && *limit < n {
		n = *limit
	}

	for i := range n {
		fmt.Println(versions[i])
	}
	return nil
}
