package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runCmd builds and runs the checksum-check binary against root, returning
// (exitCode, combinedOutput).
func runCmd(t *testing.T, root string, extraArgs ...string) (int, string) {
	t.Helper()
	args := append([]string{"run", ".", "-root", root}, extraArgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = mustGetwd(t)
	out, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), string(out)
	}
	if err != nil {
		t.Fatalf("unexpected exec error: %v\noutput:\n%s", err, out)
	}
	return 0, string(out)
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return wd
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestChecksumCheck_MatchPasses(t *testing.T) {
	root := t.TempDir()
	pack := "slug: foo\nversion: 1\n"
	writeFile(t, filepath.Join(root, "packs", "foo", "pack.yaml"), pack)
	writeFile(t, filepath.Join(root, "registry.yaml"),
		"packs:\n  - slug: foo\n    sha256: "+sha256Hex(pack)+"\n")

	code, out := runCmd(t, root)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\noutput:\n%s", code, out)
	}
}

func TestChecksumCheck_MismatchFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "packs", "foo", "pack.yaml"), "slug: foo\nversion: 1\n")
	// Wrong sha256.
	writeFile(t, filepath.Join(root, "registry.yaml"),
		"packs:\n  - slug: foo\n    sha256: "+sha256Hex("not-the-real-content")+"\n")

	code, out := runCmd(t, root)
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0\noutput:\n%s", out)
	}
	if !contains(out, "sha256 mismatch") {
		t.Fatalf("expected 'sha256 mismatch' in output, got:\n%s", out)
	}
}

func TestChecksumCheck_MissingSha256Warns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "packs", "foo", "pack.yaml"), "slug: foo\nversion: 1\n")
	writeFile(t, filepath.Join(root, "registry.yaml"),
		"packs:\n  - slug: foo\n")

	code, out := runCmd(t, root)
	if code != 0 {
		t.Fatalf("non-strict mode must not fail on missing sha256, got %d\noutput:\n%s", code, out)
	}
	if !contains(out, "WARN") {
		t.Fatalf("expected WARN in output, got:\n%s", out)
	}
}

func TestChecksumCheck_StrictModeFailsOnMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "packs", "foo", "pack.yaml"), "slug: foo\nversion: 1\n")
	writeFile(t, filepath.Join(root, "registry.yaml"),
		"packs:\n  - slug: foo\n")

	code, out := runCmd(t, root, "-strict")
	if code == 0 {
		t.Fatalf("strict mode must fail on missing sha256\noutput:\n%s", out)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
