// Command regex-check walks every packs/*/pack.yaml in the repo and
// compiles each pattern's regex (and exclude_pattern) using Go's
// RE2-based regexp package. AJV validation cannot detect Go RE2
// incompatibilities such as lookaheads, so this check guards against
// patterns that pass AJV but blow up at runtime in apps that use Go
// to compile rule regexes.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

type pattern struct {
	Name           string `yaml:"name"`
	RuleID         string `yaml:"rule_id"`
	Regex          string `yaml:"regex"`
	ExcludePattern string `yaml:"exclude_pattern"`
}

type pack struct {
	Slug     string    `yaml:"slug"`
	Patterns []pattern `yaml:"patterns"`
}

func main() {
	root := flag.String("root", ".", "registry repository root")
	flag.Parse()

	matches, err := filepath.Glob(filepath.Join(*root, "packs", "*", "pack.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob packs/*/pack.yaml: %v\n", err)
		os.Exit(2)
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "no pack.yaml files found")
		os.Exit(2)
	}

	var failures []string
	for _, file := range matches {
		failures = append(failures, checkPack(file)...)
	}

	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintln(os.Stderr, f)
		}
		fmt.Fprintf(os.Stderr, "\nRE2 compile check failed: %d invalid pattern(s)\n", len(failures))
		os.Exit(1)
	}
	fmt.Printf("RE2 compile check passed: %d pack(s)\n", len(matches))
}

func checkPack(file string) []string {
	data, err := os.ReadFile(file)
	if err != nil {
		return []string{fmt.Sprintf("%s: read: %v", file, err)}
	}

	var p pack
	if err := yaml.Unmarshal(data, &p); err != nil {
		return []string{fmt.Sprintf("%s: yaml: %v", file, err)}
	}

	var failures []string
	for _, pat := range p.Patterns {
		if pat.Regex != "" {
			if _, err := regexp.Compile(pat.Regex); err != nil {
				failures = append(failures, fmt.Sprintf(
					"%s: pack=%s rule_id=%s regex does not compile under Go RE2: %v",
					file, p.Slug, pat.RuleID, err,
				))
			}
		}
		if pat.ExcludePattern != "" {
			if _, err := regexp.Compile(pat.ExcludePattern); err != nil {
				failures = append(failures, fmt.Sprintf(
					"%s: pack=%s rule_id=%s exclude_pattern does not compile under Go RE2: %v",
					file, p.Slug, pat.RuleID, err,
				))
			}
		}
	}
	return failures
}
