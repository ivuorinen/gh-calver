// Package cmd implements the gh-calver CLI using only the Go standard library.
// Each subcommand is a *flag.FlagSet so flags are local to each command.
package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `gh-calver — Calendar Versioning for GitHub

Usage:
  gh calver <command> [flags]

Commands:
  current    Show the latest CalVer tag
  next       Preview the next CalVer tag (nothing is created)
  release    Create the next CalVer tag and publish a GitHub release
  bump       Create and push the next CalVer tag (no release)
  list       List all CalVer tags, newest first

Global flags:
  --help     Show help for any command

Run "gh calver <command> --help" for command-specific flags.`

// Execute is the main entrypoint. It dispatches to the right subcommand.
// Returns nil on success or for --help; returns an error on failure.
func Execute() error {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		return nil
	}

	sub := os.Args[1]
	args := os.Args[2:]

	var err error
	switch sub {
	case "current":
		err = runCurrent(args)
	case "next":
		err = runNext(args)
	case "release":
		err = runRelease(args)
	case "bump":
		err = runBump(args)
	case "list", "ls":
		err = runList(args)
	case "--help", "-h", "help":
		fmt.Println(usage)
		return nil
	case "--version", "version":
		fmt.Println("gh-calver dev")
		return nil
	default:
		return fmt.Errorf("unknown command %q — run \"gh calver --help\" for usage", sub)
	}

	// flag.ErrHelp means --help was requested; already printed, exit cleanly.
	if err == flag.ErrHelp {
		return nil
	}
	return err
}

// newFlagSet returns a *flag.FlagSet that writes usage to stderr on error.
func newFlagSet(name, help string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() {
		lines := strings.SplitSeq(strings.TrimSpace(help), "\n")
		for l := range lines {
			fmt.Fprintln(os.Stderr, l)
		}
	}
	return fs
}
