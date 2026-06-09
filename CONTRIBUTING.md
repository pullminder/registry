# Contributing to Pullminder Registry

Thank you for your interest in contributing detection and policy packs to the Pullminder Registry. This guide explains how packs are structured, how to create one, and how to submit it for review.

## Pack Structure

Each pack lives in its own directory under `packs/<slug>/` and must contain a `pack.yaml` file. The `pack.yaml` file defines everything about the pack: its metadata, detection patterns, scoring rules, and optional overrides.

### pack.yaml Schema

```yaml
slug: my-check                 # Unique identifier (lowercase, hyphens only)
name: My Check                 # Human-readable name
kind: detection                # Pack kind: "detection" or "policy"
action: warn                   # Action: "flag", "warn", or "block"
version: 3                     # Integer version of the pack
schema_version: 1              # Schema version (currently 1)
author: pullminder             # Author name or GitHub handle
max_weight: 10                 # Maximum weight this pack can contribute to overall score

scoring:                       # Tiered scoring thresholds
  - min_findings: 1
    score: 5
  - min_findings: 3
    score: 10

patterns:                      # Detection patterns (required for detection packs)
  - name: AWS Access Key
    rule_id: SEC-001
    regex: "AKIA[0-9A-Z]{16}"
    language: "*"              # Language filter: "*" for all, or specific like "go", "python"
    severity: critical         # Severity: "critical", "error", "high", "medium", "low", "info"
    category: security
    description: >
      AWS access keys should never appear in source code.
      Use environment variables or a secrets manager instead.
    fix_templates:             # Suggested fixes keyed by language
      go: 'os.Getenv("AWS_ACCESS_KEY_ID")'
      python: 'os.environ["AWS_ACCESS_KEY_ID"]'
      javascript: 'process.env.AWS_ACCESS_KEY_ID'

  - name: Generic Password Assignment
    rule_id: SEC-002
    regex: "(?i)(password|passwd|pwd)\\s*[:=]\\s*[\"'][^\"']{4,}[\"']"
    language: "*"
    severity: high
    category: security
    fix_templates:
      go: 'os.Getenv("DB_PASSWORD")'
      python: 'os.environ["DB_PASSWORD"]'
      javascript: 'process.env.DB_PASSWORD'

overrides:
  ignore_paths:
    - "**/*_test.go"
    - "**/testdata/**"
    - "**/fixtures/**"
```

### Top-Level Fields

| Field            | Type    | Required | Description                                                                                               |
| ---------------- | ------- | -------- | --------------------------------------------------------------------------------------------------------- |
| `slug`           | string  | Yes      | Unique pack identifier. Lowercase letters, numbers, and hyphens only.                                     |
| `name`           | string  | Yes      | Human-readable display name.                                                                              |
| `kind`           | string  | Yes      | Either `detection` (pattern matching against diffs) or `policy` (workflow rules).                         |
| `action`         | string  | Yes      | Default action when findings are produced: `flag` (reviewer brief only), `warn` (inline PR comments), or `block` (status check failure). |
| `version`        | integer | Yes      | Integer version number (e.g., `3`). Increment each time you modify patterns or configuration.             |
| `schema_version` | integer | Yes      | Must be `1` for the current schema.                                                                       |
| `author`         | string  | Yes      | Author name or GitHub handle.                                                                             |
| `max_weight`     | integer | No       | Maximum weight a single finding from this pack can contribute to the risk score (1-100). Defaults to 10.  |
| `scoring`        | array   | No       | Tiered scoring thresholds. Each entry defines `min_findings` (int) and `score` (number).                  |
| `patterns`       | array   | Yes*     | List of detection patterns. Required for `detection` packs.                                               |
| `overrides`      | object  | No       | Optional overrides such as `ignore_paths` and `ignore_authors`.                                           |

### Pattern Fields

Each entry in the `patterns` array defines a single detection rule.

| Field            | Type   | Required | Description                                                                                     |
| ---------------- | ------ | -------- | ----------------------------------------------------------------------------------------------- |
| `name`           | string | Yes      | Human-readable name of the pattern.                                                             |
| `rule_id`        | string | Yes      | Unique rule identifier within the pack (e.g., `SEC-001`). Convention: `UPPER_SNAKE_CASE`.       |
| `regex`          | string | Yes      | Regular expression to match against file contents. Must be valid RE2 syntax (Go-compatible).    |
| `language`       | string | Yes      | Language filter. Use `*` for all languages, or a specific language like `go`, `python`, `rust`. |
| `severity`       | string | Yes      | Impact level: `critical`, `error`, `high`, `medium`, `low`, or `info`.                          |
| `category`       | string | Yes      | Category grouping (e.g., `security`, `credentials`, `injection`, `code-quality`).               |
| `description`    | string | No       | Longer explanation of what the pattern detects and why it matters.                              |
| `fix_templates`  | object | No       | Suggested fixes keyed by language (e.g., `go: ...`, `python: ...`).                              |
| `exclude_pattern`| string | No       | Regex pattern to exclude false positives within matched lines.                                  |
| `path_patterns`  | array  | No       | Glob patterns restricting which file paths the rule applies to. Overrides language-based gating. |

## Creating a New Pack

1. **Fork this repository** and clone your fork locally.

2. **Create a pack directory** under `packs/` using your pack slug:

   ```
   packs/
     my-pack/
       pack.yaml
   ```

3. **Write your pack.yaml** following the schema above. Start with a small set of patterns and expand over time.

4. **Add your pack to registry.yaml** in the root of the repository. Add an entry to the `packs` array, including the sha256 of your `pack.yaml`:

   ```yaml
   packs:
     - slug: my-pack
       name: My Pack
       version: 1
       kind: detection
       default: false
       sha256: <output of `sha256sum packs/my-pack/pack.yaml`>
   ```

   The `sha256` field lets the Pullminder API reject any `pack.yaml` content
   that has been tampered with in transit. **Every time you edit `pack.yaml`,**
   recompute its sha256 and update the value in `registry.yaml` in the same commit:

   ```bash
   sha256sum packs/<slug>/pack.yaml | awk '{print $1}'
   ```

5. **Verify locally** before pushing:

   ```bash
   pullminder registry validate --strict
   ```

   This compiles every regex (Go RE2), verifies each pack's `sha256`, and runs
   schema and duplicate checks — the same validation CI performs.

6. **Submit a pull request** with your changes. CI re-runs the same checks.

## Review Expectations

When submitting a pack, reviewers will check the following:

- **Regex must compile.** All regular expressions must be valid RE2 syntax. Invalid regex will cause CI to fail.

- **No false positives on common code.** Patterns should not match typical, non-problematic code. For example, a secrets detector should not flag the word "password" in a comment or documentation string without an actual value.

- **Severity must match impact.** A `critical` severity should be reserved for issues that represent an immediate security risk (e.g., leaked credentials). Informational findings should use `info` or `low`.

- **Rule IDs must be unique within the pack.** Each `rule_id` should follow the convention `PREFIX-NNN` where PREFIX is a short uppercase identifier for the pack.

- **Pack slug must be unique across the registry.** Check the existing `registry.yaml` to ensure your slug is not already taken.

## Testing

CI automatically validates all packs on every pull request:

- **Schema validation.** Every `pack.yaml` is validated against `schema/pack.schema.json`. The registry itself is validated against `schema/registry.schema.json`.

- **Regex compilation.** All regex patterns are compiled to ensure they are valid Go RE2 syntax.

- **Per-pack sha256 checksums.** `pullminder registry validate --strict` rejects any pack whose `registry.yaml` `sha256` does not match the actual `pack.yaml` content (or is missing).

- **Duplicate detection.** CI checks for duplicate slugs and rule IDs.

Run validation locally before submitting:

```bash
# Recommended: full validation including regex compile and sha256 checks
pullminder registry validate --strict

# Alternative: JSON schema validation only (what CI also runs)
npx ajv-cli validate -s schema/pack.schema.json -d "packs/*/pack.yaml" --spec=draft2020 -c ajv-formats
npx ajv-cli validate -s schema/registry.schema.json -d registry.yaml --spec=draft2020 -c ajv-formats
```

## Code of Conduct

Be respectful and constructive in all interactions. We are building a shared resource for the developer community.
