// Command checksum-check verifies that every entry in registry.yaml whose
// "sha256" field is set matches the sha256 of the corresponding
// packs/<slug>/pack.yaml. Exits non-zero on any drift.
//
// Missing sha256 fields are reported as warnings but do not fail the build,
// so the registry can roll out checksums incrementally.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type indexFile struct {
	Packs []packMeta `yaml:"packs"`
}

type packMeta struct {
	Slug   string `yaml:"slug"`
	Sha256 string `yaml:"sha256"`
}

func main() {
	root := flag.String("root", ".", "registry repository root")
	strict := flag.Bool("strict", false, "fail when any pack lacks a sha256 (default: warn only)")
	flag.Parse()

	indexPath := filepath.Join(*root, "registry.yaml")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", indexPath, err)
		os.Exit(2)
	}

	var idx indexFile
	if err := yaml.Unmarshal(data, &idx); err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", indexPath, err)
		os.Exit(2)
	}

	if len(idx.Packs) == 0 {
		fmt.Fprintln(os.Stderr, "no packs in registry.yaml")
		os.Exit(2)
	}

	var failures, warnings []string
	for _, p := range idx.Packs {
		if p.Slug == "" {
			failures = append(failures, "pack with empty slug in registry.yaml")
			continue
		}
		packPath := filepath.Join(*root, "packs", p.Slug, "pack.yaml")
		fileBytes, err := os.ReadFile(packPath)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: read %s: %v", p.Slug, packPath, err))
			continue
		}
		actual := hashHex(fileBytes)

		if p.Sha256 == "" {
			warnings = append(warnings, fmt.Sprintf("%s: registry.yaml missing sha256 (actual=%s)", p.Slug, actual))
			continue
		}

		if p.Sha256 != actual {
			failures = append(failures, fmt.Sprintf(
				"%s: sha256 mismatch in registry.yaml\n  expected (registry.yaml): %s\n  actual   (%s):           %s",
				p.Slug, p.Sha256, packPath, actual,
			))
		}
	}

	sort.Strings(failures)
	sort.Strings(warnings)

	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "WARN:", w)
	}

	if *strict && len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\nstrict mode: %d pack(s) missing sha256\n", len(warnings))
		os.Exit(1)
	}

	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintln(os.Stderr, f)
		}
		fmt.Fprintf(os.Stderr, "\nchecksum check failed: %d mismatch(es)\n", len(failures))
		os.Exit(1)
	}

	fmt.Printf("checksum check passed: %d pack(s)\n", len(idx.Packs))
}

func hashHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
