package cmd

import (
	"fmt"

	"github.com/ivuorinen/gh-calver/internal/calver"
)

const currentHelp = `gh calver current — Show the latest CalVer tag in this repository

Usage:
  gh calver current [flags]

Flags:
  --help   Show this help

Examples:
  gh calver current`

func runCurrent(args []string) error {
	fs := newFlagSet("current", currentHelp)
	if err := fs.Parse(args); err != nil {
		return err
	}

	tags, err := ops.Tags()
	if err != nil {
		return err
	}

	latest := calver.Latest(tags)
	if latest == nil {
		return fmt.Errorf("no CalVer tags found in this repository")
	}

	fmt.Println(latest)
	return nil
}
