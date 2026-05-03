package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPackPassesValidPack(t *testing.T) {
	dir := t.TempDir()
	mkPack(t, dir, "good-pack", `slug: good-pack
patterns:
  - name: A
    rule_id: A
    regex: '\b[A-Z]{3}\b'
  - name: B
    rule_id: B
    regex: 'foo'
    exclude_pattern: 'bar'
`)

	failures := checkPack(filepath.Join(dir, "packs", "good-pack", "pack.yaml"))
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %v", failures)
	}
}

func TestCheckPackFailsLookahead(t *testing.T) {
	dir := t.TempDir()
	mkPack(t, dir, "bad-pack", `slug: bad-pack
patterns:
  - name: lookahead
    rule_id: BAD
    regex: '(?!foo)bar'
`)

	failures := checkPack(filepath.Join(dir, "packs", "bad-pack", "pack.yaml"))
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d: %v", len(failures), failures)
	}
}

func TestCheckPackFailsBadExcludePattern(t *testing.T) {
	dir := t.TempDir()
	mkPack(t, dir, "bad-exclude", `slug: bad-exclude
patterns:
  - name: ok
    rule_id: OK
    regex: 'foo'
    exclude_pattern: '(?<=bar)'
`)

	failures := checkPack(filepath.Join(dir, "packs", "bad-exclude", "pack.yaml"))
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d: %v", len(failures), failures)
	}
}

func mkPack(t *testing.T, root, slug, body string) {
	t.Helper()
	dir := filepath.Join(root, "packs", slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pack.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
