package calver_test

import (
	"testing"
	"time"

	"github.com/ivuorinen/gh-calver/internal/calver"
)

func TestParse(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
		year    int
		month   int
		micro   int
		prefix  string
	}{
		{"2025.03.0", false, 2025, 3, 0, ""},
		{"2025.03.12", false, 2025, 3, 12, ""},
		{"v2025.03.0", false, 2025, 3, 0, "v"},
		{"2025.13.0", true, 0, 0, 0, ""},  // invalid month
		{"2025.03", true, 0, 0, 0, ""},    // missing micro
		{"not-a-version", true, 0, 0, 0, ""},
		{"2025.00.0", true, 0, 0, 0, ""},  // month 0
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			v, err := calver.Parse(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Year != tc.year || v.Month != tc.month || v.Micro != tc.micro || v.Prefix != tc.prefix {
				t.Errorf("got %+v, want year=%d month=%d micro=%d prefix=%q",
					v, tc.year, tc.month, tc.micro, tc.prefix)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	v := &calver.Version{Year: 2025, Month: 3, Micro: 0}
	if got := v.String(); got != "2025.03.0" {
		t.Errorf("got %q, want %q", got, "2025.03.0")
	}

	v2 := &calver.Version{Year: 2025, Month: 3, Micro: 5, Prefix: "v"}
	if got := v2.String(); got != "v2025.03.5" {
		t.Errorf("got %q, want %q", got, "v2025.03.5")
	}
}

func TestLatest(t *testing.T) {
	tags := []string{"2025.01.0", "2025.03.2", "2025.03.0", "2025.03.1", "not-calver", "2024.12.5"}
	got := calver.Latest(tags)
	if got == nil {
		t.Fatal("got nil, expected a version")
	}
	if got.String() != "2025.03.2" {
		t.Errorf("got %s, want 2025.03.2", got)
	}
}

func TestLatestEmpty(t *testing.T) {
	if got := calver.Latest(nil); got != nil {
		t.Errorf("expected nil for empty tags, got %v", got)
	}
	if got := calver.Latest([]string{"not-a-tag", "also-not"}); got != nil {
		t.Errorf("expected nil for no calver tags, got %v", got)
	}
}

func TestNext_NewPeriod(t *testing.T) {
	// No existing tags for 2025.03 — should start at 0
	tags := []string{"2025.02.0", "2025.02.1", "2024.12.3"}
	now := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	got := calver.Next(tags, now, "")
	if got.String() != "2025.03.0" {
		t.Errorf("got %s, want 2025.03.0", got)
	}
}

func TestNext_SamePeriodCollision(t *testing.T) {
	// 2025.03.0 and 2025.03.1 already exist — should produce 2025.03.2
	tags := []string{"2025.03.0", "2025.03.1", "2025.02.0"}
	now := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	got := calver.Next(tags, now, "")
	if got.String() != "2025.03.2" {
		t.Errorf("got %s, want 2025.03.2", got)
	}
}

func TestNext_WithPrefix(t *testing.T) {
	tags := []string{"v2025.03.0"}
	now := time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)
	got := calver.Next(tags, now, "v")
	if got.String() != "v2025.03.1" {
		t.Errorf("got %s, want v2025.03.1", got)
	}
}

func TestNext_NoTags(t *testing.T) {
	now := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	got := calver.Next(nil, now, "")
	if got.String() != "2025.03.0" {
		t.Errorf("got %s, want 2025.03.0", got)
	}
}

func TestAll(t *testing.T) {
	tags := []string{"2025.01.0", "2025.03.1", "2024.12.5", "2025.03.0", "bad-tag"}
	versions := calver.All(tags)
	if len(versions) != 4 {
		t.Fatalf("expected 4 versions, got %d", len(versions))
	}
	if versions[0].String() != "2025.03.1" {
		t.Errorf("first should be 2025.03.1, got %s", versions[0])
	}
}

func TestFilter(t *testing.T) {
	tags := []string{"2025.03.0", "v2025.03.1", "not-semver", "1.2.3", "2025.03.2"}
	got := calver.Filter(tags)
	if len(got) != 3 {
		t.Errorf("expected 3 calver tags, got %d", len(got))
	}
}

func TestVersionCore(t *testing.T) {
	v := &calver.Version{Year: 2025, Month: 3, Micro: 7, Prefix: "v"}
	// Core() must strip the prefix.
	if got := v.Core(); got != "2025.03.7" {
		t.Errorf("Core() = %q, want %q", got, "2025.03.7")
	}
}

func TestVersionCore_NoPrefix(t *testing.T) {
	v := &calver.Version{Year: 2025, Month: 12, Micro: 0}
	if got := v.Core(); got != "2025.12.0" {
		t.Errorf("Core() = %q, want %q", got, "2025.12.0")
	}
}

func TestVersionEqual(t *testing.T) {
	a := &calver.Version{Year: 2025, Month: 3, Micro: 0}
	b := &calver.Version{Year: 2025, Month: 3, Micro: 0, Prefix: "v"}
	if !a.Equal(b) {
		t.Errorf("Equal() should be true when YYYY.MM.MICRO matches regardless of prefix")
	}
}

func TestVersionEqual_Different(t *testing.T) {
	cases := []struct{ a, b *calver.Version }{
		{
			&calver.Version{Year: 2025, Month: 3, Micro: 0},
			&calver.Version{Year: 2025, Month: 3, Micro: 1},
		},
		{
			&calver.Version{Year: 2025, Month: 3, Micro: 0},
			&calver.Version{Year: 2025, Month: 4, Micro: 0},
		},
		{
			&calver.Version{Year: 2025, Month: 3, Micro: 0},
			&calver.Version{Year: 2024, Month: 3, Micro: 0},
		},
	}
	for _, tc := range cases {
		if tc.a.Equal(tc.b) {
			t.Errorf("Equal() should be false: %v vs %v", tc.a, tc.b)
		}
	}
}
