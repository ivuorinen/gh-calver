package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ivuorinen/gh-calver/internal/git"
)

// ── mock git.Client ───────────────────────────────────────────────────────────

// mockOps is a configurable mock implementing git.Client.
// Fields starting with "On" are called by the corresponding method.
// If a field is nil the method returns a zero-value response.
type mockOps struct {
	OnTags          func() ([]string, error)
	OnTagExists     func(tag string) bool
	OnCreateTag     func(tag, message string) error
	OnPushTag       func(tag string) error
	OnDeleteTag     func(tag string) error
	OnCreateRelease func(tag string, opts git.ReleaseOptions) error

	// Recorded calls for assertions.
	CreatedTags  []string
	PushedTags   []string
	DeletedTags  []string
	ReleasedTags []string
	ReleaseOpts  []git.ReleaseOptions
}

func (m *mockOps) Tags() ([]string, error) {
	if m.OnTags != nil {
		return m.OnTags()
	}
	return nil, nil
}

func (m *mockOps) TagExists(tag string) bool {
	if m.OnTagExists != nil {
		return m.OnTagExists(tag)
	}
	return false
}

func (m *mockOps) CreateTag(tag, message string) error {
	m.CreatedTags = append(m.CreatedTags, tag)
	if m.OnCreateTag != nil {
		return m.OnCreateTag(tag, message)
	}
	return nil
}

func (m *mockOps) PushTag(tag string) error {
	m.PushedTags = append(m.PushedTags, tag)
	if m.OnPushTag != nil {
		return m.OnPushTag(tag)
	}
	return nil
}

func (m *mockOps) DeleteTag(tag string) error {
	m.DeletedTags = append(m.DeletedTags, tag)
	if m.OnDeleteTag != nil {
		return m.OnDeleteTag(tag)
	}
	return nil
}

func (m *mockOps) CreateRelease(tag string, opts git.ReleaseOptions) error {
	m.ReleasedTags = append(m.ReleasedTags, tag)
	m.ReleaseOpts = append(m.ReleaseOpts, opts)
	if m.OnCreateRelease != nil {
		return m.OnCreateRelease(tag, opts)
	}
	return nil
}

// ── stdout capture ────────────────────────────────────────────────────────────

// captureStdout runs f and returns everything written to os.Stdout.
func captureStdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("os.Pipe: %v", err))
	}
	old := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// ── time stub ─────────────────────────────────────────────────────────────────

// fixedTime pins the clock to a deterministic value for all cmd tests.
var testTime = time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)

// withMock installs a mock client + fixed clock and restores them after the test.
// Returns the mock so callers can configure OnXxx handlers before use.
func withMock(m *mockOps) func() {
	origOps := ops
	origNow := nowFn
	ops = m
	nowFn = func() time.Time { return testTime }
	return func() {
		ops = origOps
		nowFn = origNow
	}
}

// errSentinel is a reusable error for testing failure paths in ops.
var errSentinel = errors.New("sentinel error")
