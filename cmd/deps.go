package cmd

import (
	"time"

	"github.com/ivuorinen/gh-calver/internal/git"
)

// ops is the git/gh client used by every subcommand.
// Tests replace this with a mock; production code uses the real client.
var ops git.Client = git.NewClient()

// nowFn returns the current time. Tests can replace it to control version output.
var nowFn = time.Now
