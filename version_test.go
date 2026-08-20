package paubox

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// changelogRelease matches a released version heading in CHANGELOG.md, e.g.
// "## [1.0.0] - 2026-08-20". The "## [Unreleased]" heading carries no version
// number and so never matches.
var changelogRelease = regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`)

func TestVersionMatchesChangelog(t *testing.T) {
	data, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}

	m := changelogRelease.FindSubmatch(data)
	if m == nil {
		t.Fatal("CHANGELOG.md has no released version heading")
	}

	if got := string(m[1]); got != Version {
		t.Errorf("newest CHANGELOG.md release is %s but Version is %s; bump both together", got, Version)
	}
}

func TestDefaultUserAgentEmbedsVersion(t *testing.T) {
	if want := "paubox-go/" + Version; defaultUserAgent != want {
		t.Errorf("defaultUserAgent = %q, want %q", defaultUserAgent, want)
	}
}

// TestVersionMatchesReleaseTag is the guard the release workflow runs with
// RELEASE_TAG set to the tag being published. A tag that disagrees with
// Version would ship an SDK that misreports itself in the User-Agent header.
func TestVersionMatchesReleaseTag(t *testing.T) {
	tag := os.Getenv("RELEASE_TAG")
	if tag == "" {
		t.Skip("RELEASE_TAG unset; this check runs in the release workflow")
	}

	if want := "v" + Version; tag != want {
		t.Errorf("release tag is %s but Version is %s; expected tag %s", tag, Version, want)
	}
}

func TestVersionIsSemver(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`).MatchString(Version) {
		t.Errorf("Version = %q is not a valid semantic version", Version)
	}
}

// TestModulePathMatchesMajorVersion enforces the Go rule that a v2+ module
// carries a /vN suffix on its module path while v0 and v1 must not.
//
// This exists because release-please bumps Version from a `feat!:` commit but
// never edits go.mod. Without this check a breaking change would produce a
// v2.0.0 tag on a module still declaring itself v1 — a release the module
// proxy refuses to resolve. The fix when this fails is to update the module
// path in go.mod and every internal import in the same pull request.
func TestModulePathMatchesMajorVersion(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	m := regexp.MustCompile(`(?m)^module\s+(\S+)`).FindSubmatch(data)
	if m == nil {
		t.Fatal("go.mod has no module directive")
	}
	modulePath := string(m[1])

	major, _, _ := strings.Cut(Version, ".")
	hasSuffix := regexp.MustCompile(`/v\d+$`).MatchString(modulePath)

	if major == "0" || major == "1" {
		if hasSuffix {
			t.Errorf("module path %q carries a major-version suffix, which is invalid at v%s", modulePath, major)
		}
		return
	}

	if want := "/v" + major; !strings.HasSuffix(modulePath, want) {
		t.Errorf("Version is %s but module path is %q; go.mod must end in %s", Version, modulePath, want)
	}
}
