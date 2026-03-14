# gh-calver

A [GitHub CLI](https://cli.github.com/) extension for [Calendar Versioning](https://calver.org/)
(CalVer) release management.

Versions use the `YYYY.MM.MICRO` format (e.g. `2025.03.0`):

| Segment | Meaning |
|---------|---------|
| `YYYY`  | Full four-digit year |
| `MM`    | Zero-padded two-digit month |
| `MICRO` | Build counter — resets to `0` on a new `YYYY.MM`, increments on collision |

## Installation

```sh
gh extension install ivuorinen/gh-calver
```

Requires [gh](https://cli.github.com/) ≥ 2.x and Git.

## Commands

### `gh calver current`

Print the latest CalVer tag in the repository.

```sh
gh calver current
# 2025.03.4
```

Exits with status 1 if no CalVer tags exist.

---

### `gh calver next`

Preview the next CalVer version without creating anything.

```sh
gh calver next
# 2025.03.5

gh calver next --prefix v
# v2025.03.5
```

---

### `gh calver release`

Create the next CalVer git tag, push it, and publish a GitHub release.

```sh
# Minimal — uses GitHub auto-generated release notes
gh calver release

# With a v-prefix
gh calver release --prefix v

# Custom release notes from a file
gh calver release --notes-file CHANGELOG.md

# Draft pre-release with a custom title
gh calver release --draft --prerelease --title "March 2025 beta"

# Attach binary assets to the release
gh calver release dist/app-linux-amd64.tar.gz dist/app-darwin-arm64.tar.gz

# Preview without doing anything
gh calver release --dry-run
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--prefix` | `""` | Tag prefix, e.g. `v` |
| `--title` | version string | Release title |
| `--notes-file` | — | Path to Markdown file for release notes |
| `--target` | HEAD | Branch or SHA to tag |
| `--message` | — | Annotated tag message (lightweight tag if omitted) |
| `--draft` | false | Save as draft |
| `--prerelease` | false | Mark as pre-release |
| `--dry-run` | false | Print what would happen, do nothing |

---

### `gh calver bump`

Create and push a CalVer tag without creating a GitHub release.

```sh
gh calver bump
gh calver bump --prefix v --message "Release 2025.03.0"
gh calver bump --dry-run
```

Useful as part of a custom CI/CD pipeline where you handle release creation
separately (e.g. with GoReleaser).

---

### `gh calver list`

List all CalVer tags in the repository, sorted newest-first.

```sh
gh calver list
gh calver list --limit 5
```

---

## How MICRO increments

| Situation | Result |
|-----------|--------|
| No tags exist at all | `2025.03.0` |
| Tags exist, but none for the current month | `2025.03.0` |
| `2025.03.0` already exists | `2025.03.1` |
| `2025.03.0` and `2025.03.1` already exist | `2025.03.2` |

## v-prefix support

All commands accept a `--prefix` flag. If your existing tags use `v`, pass
`--prefix v` to each command. There is no global config file — this is
intentional to keep the tool simple and scriptable.

```sh
# In a Makefile or CI script
NEXT=$(gh calver next --prefix v)
gh calver release --prefix v --notes-file CHANGELOG.md
```

## Development

```sh
git clone https://github.com/ivuorinen/gh-calver
cd gh-calver

make test       # run unit tests
make lint       # go vet + staticcheck
make build      # build for the current platform
make build-all  # cross-compile for linux/darwin/windows × amd64/arm64
make install    # install as a local gh extension
make release    # tag + push + goreleaser (from clean main only)
```

### Releasing

`make release` automates the full release flow:

1. Verifies the working tree is clean (no uncommitted changes)
2. Verifies you are on the `main` branch
3. Computes the next CalVer tag via `go run . next`
4. Creates an annotated git tag
5. Pushes the tag to origin
6. Runs [GoReleaser](https://goreleaser.com/) to build and publish the release

```sh
# From a clean main branch
make release
```

### Installing locally

```sh
# From inside the cloned repo
gh extension install .
```

## License

MIT
