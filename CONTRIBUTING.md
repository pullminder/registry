# Contributing to Pullminder Registry

Thank you for your interest in contributing detection and policy packs to the Pullminder Registry. This guide explains how packs are structured, how to create one, and how to submit it for review.

## Pack Structure

Each pack lives in its own directory under `packs/<slug>/` and must contain a `pack.yaml` file. The `pack.yaml` file defines everything about the pack: its metadata, detection patterns, scoring rules, and optional overrides.

### pack.yaml Schema

```yaml
slug: secrets                    # Unique identifier (lowercase, hyphens only)
name: Secrets Detection          # Human-readable name
kind: detection                  # Pack kind: "detection" or "policy"
action: flag                     # Action: "flag", "block", or "warn"
version: 1.0.0                  # Semantic version of the pack
schema_version: 1               # Schema version (currently 1)
author: pullminder               # Author name or GitHub handle
max_weight: 10                   # Maximum weight this pack can contribute to overall score

scoring:
  model: additive                # Scoring model: "additive" or "weighted"
  base_weight: 5                 # Base weight when any pattern matches

patterns:
  - name: AWS Access Key
    rule_id: SEC-001
    regex: "AKIA[0-9A-Z]{16}"
    language: "*"                # Language filter: "*" for all, or specific like "go", "python"
    severity: critical           # Severity: "critical", "high", "medium", "low", "info"
    category: credentials
    fix_templates:
      - "Move this credential to an environment variable"
      - "Use a secrets manager such as AWS Secrets Manager or HashiCorp Vault"

  - name: Generic Password Assignment
    rule_id: SEC-002
    regex: "(?i)(password|passwd|pwd)\\s*[:=]\\s*[\"'][^\"']{4,}[\"']"
    language: "*"
    severity: high
    category: credentials
    fix_templates:
      - "Replace hardcoded password with an environment variable reference"

overrides:
  ignore_paths:
    - "**/*_test.go"
    - "**/testdata/**"
    - "**/fixtures/**"
```

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `slug` | string | Yes | Unique pack identifier. Lowercase letters and hyphens only. |
| `name` | string | Yes | Human-readable display name. |
| `kind` | string | Yes | Either `detection` (finds patterns in code) or `policy` (enforces workflow rules). |
| `action` | string | Yes | What happens on match: `flag` (add comment), `block` (request changes), or `warn` (add warning). |
| `version` | string | Yes | Semantic version of the pack (e.g., `1.0.0`). |
| `schema_version` | integer | Yes | Must be `1` for the current schema. |
| `author` | string | Yes | Author name or GitHub handle. |
| `max_weight` | integer | Yes | Maximum score weight this pack can contribute (1-10). |
| `scoring` | object | Yes | Scoring configuration with `model` and `base_weight`. |
| `patterns` | array | Yes | List of detection patterns (see below). |
| `overrides` | object | No | Optional overrides such as `ignore_paths`. |

### Pattern Fields

Each entry in the `patterns` array defines a single detection rule.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable name of the pattern. |
| `rule_id` | string | Yes | Unique rule identifier within the pack (e.g., `SEC-001`). |
| `regex` | string | Yes | Regular expression to match against file contents. Must be valid RE2 syntax. |
| `language` | string | Yes | Language filter. Use `*` for all languages, or a specific language like `go`, `python`, `rust`. |
| `severity` | string | Yes | Impact level: `critical`, `high`, `medium`, `low`, or `info`. |
| `category` | string | Yes | Category grouping (e.g., `credentials`, `injection`, `misconfiguration`). |
| `fix_templates` | array | No | List of suggested fix descriptions shown to the developer. |

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
   that has been tampered with in transit. Anyone who edits `pack.yaml` must
   recompute and update this value in the same commit.

5. **Verify locally** before pushing:

   ```bash
   go run ./cmd/checksum-check -root . -strict
   go run ./cmd/regex-check    -root .
   ```

6. **Submit a pull request** with your changes. CI re-runs both checks.

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

- **Regex compilation.** All regex patterns are compiled to ensure they are valid.

- **Per-pack sha256 checksums.** `cmd/checksum-check -strict` rejects any pack whose `registry.yaml` `sha256` does not match the actual `pack.yaml` content (or is missing).

- **Duplicate detection.** CI checks for duplicate slugs and rule IDs.

To run validation locally before submitting:

```bash
npm install -g ajv-cli ajv-formats
ajv validate -s schema/pack.schema.json -d "packs/*/pack.yaml" --spec=draft2020 -c ajv-formats
ajv validate -s schema/registry.schema.json -d registry.yaml --spec=draft2020 -c ajv-formats
```

## Code of Conduct

Be respectful and constructive in all interactions. We are building a shared resource for the developer community.
