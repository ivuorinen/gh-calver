// Package calver provides Calendar Versioning (CalVer) parsing and calculation
// using the YYYY.MM.MICRO format (e.g. 2025.03.0).
package calver

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// versionRe matches optional v-prefix + YYYY.MM.MICRO
var versionRe = regexp.MustCompile(`^(v?)(\d{4})\.(\d{2})\.(\d+)$`)

// Version represents a single CalVer release.
type Version struct {
	Year   int
	Month  int
	Micro  int
	Prefix string // "v" or ""
}

// String returns the version in YYYY.MM.MICRO format (with prefix if set).
func (v *Version) String() string {
	return fmt.Sprintf("%s%d.%02d.%d", v.Prefix, v.Year, v.Month, v.Micro)
}

// Core returns the version without prefix.
func (v *Version) Core() string {
	return fmt.Sprintf("%d.%02d.%d", v.Year, v.Month, v.Micro)
}

// Equal returns true if two versions refer to the same release.
func (v *Version) Equal(other *Version) bool {
	return v.Year == other.Year && v.Month == other.Month && v.Micro == other.Micro
}

// Parse parses a calver string. Accepts optional leading "v".
func Parse(s string) (*Version, error) {
	m := versionRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return nil, fmt.Errorf("not a valid YYYY.MM.MICRO calver tag: %q", s)
	}
	year, _ := strconv.Atoi(m[2])
	month, _ := strconv.Atoi(m[3])
	micro, _ := strconv.Atoi(m[4])

	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month %d in calver tag: %q", month, s)
	}
	return &Version{Prefix: m[1], Year: year, Month: month, Micro: micro}, nil
}

// Filter returns only the tags that parse as valid calver versions.
func Filter(tags []string) []*Version {
	out := make([]*Version, 0, len(tags))
	for _, t := range tags {
		if v, err := Parse(t); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// Sort sorts a slice of versions newest-first (Year desc, Month desc, Micro desc).
func Sort(versions []*Version) {
	slices.SortFunc(versions, func(a, b *Version) int {
		if c := cmp.Compare(b.Year, a.Year); c != 0 {
			return c
		}
		if c := cmp.Compare(b.Month, a.Month); c != 0 {
			return c
		}
		return cmp.Compare(b.Micro, a.Micro)
	})
}

// Latest returns the most recent version from a list of tag strings.
// Returns nil if no valid calver tags are found.
func Latest(tags []string) *Version {
	versions := Filter(tags)
	if len(versions) == 0 {
		return nil
	}
	Sort(versions)
	return versions[0]
}

// Next calculates the next CalVer version given existing tags and the current
// time. MICRO resets to 0 on a new YYYY.MM, and increments by 1 when the
// current period already has a release.
func Next(tags []string, now time.Time, prefix string) *Version {
	year := now.Year()
	month := int(now.Month())

	micro := 0
	for _, v := range Filter(tags) {
		if v.Year == year && v.Month == month && v.Micro >= micro {
			micro = v.Micro + 1
		}
	}

	return &Version{Year: year, Month: month, Micro: micro, Prefix: prefix}
}

// All returns all calver versions from tag list, sorted newest-first.
func All(tags []string) []*Version {
	versions := Filter(tags)
	Sort(versions)
	return versions
}
