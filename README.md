# Pullminder Registry

> Official rule pack registry for Pullminder — detection rules, policy packs, and community contributions.

[![License](https://img.shields.io/github/license/pullminder/registry)](LICENSE)

## Overview

Rule packs are bundles of detection patterns and policy checks that Pullminder runs against every pull request. Each pack targets a specific concern -- leaked secrets, language-specific vulnerabilities, workflow standards, and more. When a pattern matches a line in the PR diff, Pullminder creates a **finding** that feeds into the risk score.

## Detection packs vs policy packs

Pullminder ships two kinds of rule packs:

- **Detection packs** use regex pattern matching against the PR diff to find security issues, risky code patterns, and notable changes. Each pattern targets specific languages or file types and carries a severity level.
- **Policy packs** enforce team workflow standards such as test naming conventions, PR description quality, and commit message formatting. They evaluate structural properties of the PR rather than individual lines of code.

Both kinds of packs produce findings that contribute to the overall risk score.

## How packs are evaluated

On every pull request event (open, synchronize, reopen), Pullminder fetches the diff and runs it through all enabled packs in order:

1. Each enabled pack iterates over its patterns.
2. Each pattern is matched against the relevant files in the diff, filtered by language and path.
3. When a pattern matches, Pullminder creates a finding with the pattern's severity, category, and suggested fix.
4. Findings are scored using the pack's scoring model and aggregated into the PR risk score.

Packs that do not match any file in the diff produce no findings and have zero impact on the score.

## Actions

Every pack declares an **action** that determines what happens when it produces findings:

| Action | Behavior | Used by |
|--------|----------|---------|
| `flag` | Add findings to the risk score and include them in the reviewer brief. This is the default. | Available as an override for any pack |
| `warn` | Post an inline comment on the PR for each finding. Findings also contribute to the risk score. | Most packs (language security, governance, policy, bot-detection, ai-detection, dependency-detection, sensitive-paths) |
| `block` | Set the Pullminder status check to "failure" so the PR cannot be merged until findings are resolved. | `secrets`, `infra-security` |

You can override the action for any pack in the dashboard under **Settings > Rule packs**.

## Default packs vs optional packs

Pullminder enables a set of default packs out of the box. These cover the most common concerns and require no configuration. Optional packs target specific languages or advanced use cases and must be enabled explicitly.

## Available packs

The following packs are available from the official Pullminder registry at [github.com/pullminder/registry](https://github.com/pullminder/registry).

### Detection packs

| Slug | Name | Description | Default |
|------|------|-------------|---------|
| `secrets` | Secrets | Leaked API keys, tokens, passwords, and credentials in code | Yes |
| `go-security` | Go Security | SQL injection, command injection, and unsafe pointer usage in Go | No |
| `python-security` | Python Security | Unsafe deserialization, eval, and SQL injection in Python | No |
| `rust-security` | Rust Security | Unsafe blocks and memory safety issues in Rust | No |
| `ruby-security` | Ruby Security | Mass assignment, command injection, and common Rails vulnerabilities | No |
| `php-security` | PHP Security | SQL injection, file inclusion, and remote code execution in PHP | No |
| `react-security` | React Security | XSS vectors, unsafe innerHTML, and client-side injection in React/JSX | No |
| `infra-security` | Infrastructure Security | Dockerfile, Kubernetes, Terraform, and CI/CD misconfiguration | Yes |
| `bot-detection` | Bot Detection | PRs opened by bots and automated tools | Yes |
| `ai-detection` | AI Detection | AI-generated code detection patterns | Yes |
| `dependency-detection` | Dependency Detection | Lockfile and manifest changes (package.json, go.sum, etc.) | Yes |
| `java-security` | Java Security | SQL injection, XXE, deserialization in Java | No |
| `csharp-security` | C# Security | SqlCommand injection, BinaryFormatter in C# | No |
| `kotlin-security` | Kotlin Security | WebView, SharedPreferences, exported component issues in Kotlin/Android | No |
| `swift-security` | Swift Security | ATS bypass, keychain, biometric auth issues in Swift/iOS | No |
| `shell-security` | Shell Security | Eval injection, curl-pipe, chmod issues in Shell/Bash | No |
| `sensitive-paths` | Sensitive Paths | Changes to auth, crypto, permissions, and infrastructure files | Yes |

### Governance and cross-cutting packs

| Slug | Name | Description | Default |
|------|------|-------------|---------|
| `ai-senior-review` | AI Senior Review Required | Requires senior reviewer approval for high-risk AI-generated PRs | Yes |
| `pii-leakage` | PII Leakage Detection | Detects PII (SSN, credit cards, emails) in logging and output contexts | No |
| `crypto-anti-patterns` | Cryptographic Anti-Patterns | Language-agnostic weak crypto detection (MD5, DES, ECB, small keys) | No |
| `migration-safety` | Migration Safety | Dangerous SQL migration patterns (DROP TABLE, type changes, missing defaults) | No |
| `license-risk` | License Risk Detection | Flags copyleft license introductions (GPL, AGPL, SSPL) in dependency manifests | No |
| `owasp-mapping` | OWASP Top 10 Mapping | Maps detection rules to OWASP Top 10 categories for compliance reporting | No |

### Policy packs

| Slug | Name | Description | Default |
|------|------|-------------|---------|
| `test-conventions` | Test Conventions | Test naming, coverage thresholds, and file organization standards | Yes |
| `review-quality` | Review Quality | PR description completeness, commit message format, and review standards | Yes |

## Managing packs

### Via the CLI

```bash
# List all available packs and their status
pullminder packs list

# Enable an optional pack
pullminder packs enable go-security

# Disable a default pack
pullminder packs disable bot-detection
```

### Via the dashboard

Navigate to **Settings > Rule packs** in the Pullminder dashboard. Each pack shows its current status, the number of findings it has produced, and a toggle to enable or disable it. You can also override the action (flag, warn, block) for each pack from this view.

## Custom registries

In addition to the official registry, you can configure a custom private registry for proprietary rule packs. Custom registries use the same pack format and can be hosted as a Git repository or served from an HTTP endpoint. See the [Authoring custom packs](/packs/authoring-guide/) guide for details on creating your own packs and registries.

## Authoring Guide

Pullminder rule packs are portable YAML files that define detection patterns or policy checks. You can create custom packs for your organization's specific needs, test them locally, and publish them to the community registry or host them in a private registry.

## Prerequisites

- Pullminder CLI installed ([Installation guide](/cli/installation/))
- Familiarity with regular expressions
- A GitHub account (for publishing to the community registry)

## Step 1: Scaffold a registry

A registry is a directory that holds one or more rule packs. Start by scaffolding a new registry:

```bash
pullminder registry init my-rules
```

This creates the following structure:

```
my-rules/
  registry.yaml       # Registry metadata
  packs/              # Pack definitions go here
  README.md           # Registry documentation
```

The `registry.yaml` file identifies the registry and contains metadata used when syncing packs:

```yaml
name: my-rules
description: Custom rule packs for my organization
author: your-github-handle
url: https://github.com/your-org/my-rules
```

## Step 2: Add a pack

Use the CLI to scaffold a new pack inside the registry:

```bash
pullminder registry pack add \
  --slug my-check \
  --kind detection \
  --name "My Check"
```

This creates `packs/my-check/pack.yaml` with a minimal template.

## Step 3: Edit pack.yaml

Open `packs/my-check/pack.yaml` and configure every field. Here is the full anatomy of a pack file:

```yaml
slug: my-check
name: My Check
kind: detection
action: flag
version: 1
max_weight: 10

scoring:
  - min_findings: 1
    score: 5
  - min_findings: 3
    score: 10

patterns:
  - name: Hardcoded database password
    rule_id: MC-001
    regex: "(?i)db_password\\s*=\\s*[\"'][^\"']+[\"']"
    language: "*"
    severity: error
    category: security
    description: Database password is hardcoded in source code.

  - name: Console log in production code
    rule_id: MC-002
    regex: "console\\.log\\("
    language: javascript
    severity: low
    category: code-quality
    description: Console.log statements should be removed before merging.

overrides:
  ignore_paths:
    - "**/testdata/**"
    - "**/fixtures/**"
  ignore_authors:
    - "dependabot[bot]"
```

### Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `slug` | Yes | Unique identifier for the pack. Lowercase, hyphens only. |
| `name` | Yes | Human-readable display name. |
| `kind` | Yes | Either `detection` (pattern matching) or `policy` (workflow rules). |
| `action` | Yes | Default action when findings are produced: `flag`, `warn`, or `block`. |
| `version` | Yes | Integer version of the pack (e.g., `3`). Increment each time you modify patterns or configuration. |
| `schema_version` | No | Schema version. Omit if not needed. The registry infers schema version from the pack structure. |
| `author` | No | GitHub handle of the pack author. Required when publishing to the community registry. |
| `max_weight` | No | Maximum weight a single finding from this pack can contribute to the risk score. Defaults to `10`. |
| `scoring` | No | Tiered scoring configuration. Each tier defines the minimum number of findings required to reach that score. The pack's contribution to the risk score is the highest tier whose `min_findings` threshold is met. See [Pack schema reference](/packs/pack-reference/) for details. |
| `patterns` | Yes | Array of pattern objects. At least one pattern is required for detection packs. |
| `overrides` | No | Path and author exclusions. |

## Step 4: Write patterns

Each pattern in the `patterns` array defines a single detection rule. Patterns are matched against the added and modified lines in the PR diff.

### Regex syntax

Patterns use Go-compatible regular expressions (RE2 syntax). A few tips:

- Use `(?i)` at the start for case-insensitive matching.
- Escape special characters: `\\.`, `\\(`, `\\{`.
- Use `[^\"']+` to match non-empty strings inside quotes.
- Use `\\b` for word boundaries to avoid false positives.

### Language targeting

The `language` field filters which files the pattern runs against:

| Value | Matches |
|-------|---------|
| `*` | All files |
| `go` | `.go` files |
| `python` | `.py` files |
| `javascript` | `.js`, `.jsx`, `.mjs` files |
| `typescript` | `.ts`, `.tsx` files |
| `rust` | `.rs` files |
| `ruby` | `.rb`, `.erb` files |
| `php` | `.php` files |
| `java` | `.java` files |
| `yaml` | `.yaml`, `.yml` files |
| `dockerfile` | `Dockerfile`, `*.dockerfile` files |
| `terraform` | `.tf` files |

### Severity levels

| Severity | Weight | Meaning |
|----------|--------|---------|
| `critical` | 10 | Immediate security risk. Typically warrants blocking the PR. |
| `error` | 8 | Serious error that should be fixed before merging. |
| `high` | 7 | Serious issue that should be addressed before merging. |
| `medium` | 5 | Notable concern worth reviewing. |
| `low` | 3 | Minor issue or style violation. |
| `info` | 1 | Informational finding. Does not significantly affect the risk score. |

## Step 5: Test the pack

Run the pack against a local repository or diff to verify it works:

```bash
pullminder rules test --pack my-check --verbose
```

This runs the pack against the current working directory's latest diff and prints each matched pattern with the file, line number, and matched text.

To test against a specific diff:

```bash
pullminder rules test --pack my-check --diff ./path/to/diff.patch
```

To test against a specific file:

```bash
pullminder rules test --pack my-check --file ./src/config.py
```

## Step 6: Validate the registry

Before publishing, validate that all packs in the registry conform to the schema:

```bash
pullminder registry validate --strict
```

The `--strict` flag enables additional checks:

- Every pattern must have a unique `rule_id`.
- The `regex` field must compile without errors.
- The `version` field must be a positive integer.
- The `severity` field must be one of the allowed values.

Fix any validation errors before proceeding.

## Step 7: Publish to the community registry

To share your pack with the Pullminder community:

```bash
pullminder rules publish --pack my-check
```

This submits the pack for review. Once approved, it appears in the official registry and can be enabled by any Pullminder user from the dashboard or CLI.

### Publishing requirements

- The pack must pass `pullminder registry validate --strict`.
- The `author` field must match your authenticated GitHub handle.
- The pack must include at least one pattern with a description.
- The `slug` must not conflict with an existing pack in the registry.

## Step 8: Use as a custom private registry

If you prefer to keep packs private, push the registry to a Git repository and configure it in the Pullminder dashboard:

1. Push your registry to a private GitHub repository.
2. In the Pullminder dashboard, navigate to **Settings > Registries**.
3. Click **Add registry** and enter the repository URL.
4. Pullminder syncs the registry and makes its packs available for your organization.

You can also configure a custom registry via the CLI:

```bash
pullminder registry add https://github.com/your-org/my-rules.git
```

Pullminder pulls packs from custom registries on every sync cycle (default: every 15 minutes) and applies them alongside the official packs.

## Pack Reference

This page documents every field in a Pullminder rule pack YAML file. For a step-by-step walkthrough of creating a pack, see the [Authoring custom packs](/packs/authoring-guide/) guide.

## Full schema

```yaml
slug: string               # Required. Unique pack identifier.
name: string               # Required. Display name.
kind: detection | policy   # Required. Pack type.
action: flag | warn | block # Required. Default action on findings.
version: integer           # Required. Integer version (e.g., 3).
schema_version: integer    # Optional. Schema version.
author: string             # Optional. GitHub handle (required for publishing).
max_weight: integer        # Optional. Max weight per finding. Default: 10.

scoring:                   # Optional. Tiered scoring thresholds.
  - min_findings: integer  #   Minimum findings to reach this score.
    score: integer         #   Risk score contribution at this tier.

patterns:                  # Required for detection packs. Array of pattern objects.
  - name: string           #   Required. Human-readable pattern name.
    rule_id: string        #   Required. Unique identifier (e.g., "SEC-001").
    regex: string          #   Required. RE2-compatible regular expression.
    language: string       #   Required. Language filter ("*" for all).
    severity: string       #   Required. One of: critical, error, high, medium, low, info.
    category: string       #   Required. Freeform category (e.g., "security").
    description: string    #   Optional. Detailed explanation of the finding.
    fix_templates:         #   Optional. Array of suggested fix strings.
      - string

overrides:                 # Optional. Exclusion rules.
  ignore_paths:            #   Optional. Glob patterns for paths to skip.
    - string
  ignore_authors:          #   Optional. GitHub usernames to skip.
    - string
```

## Top-level fields

### `slug`

**Type:** string -- **Required**

Unique identifier for the pack within a registry. Must be lowercase and may contain only letters, numbers, and hyphens. This value is used in CLI commands and API calls to reference the pack.

```yaml
slug: my-custom-check
```

### `name`

**Type:** string -- **Required**

Human-readable display name shown in the dashboard, PR comments, and CLI output.

```yaml
name: My Custom Check
```

### `kind`

**Type:** enum -- **Required**

Determines how the pack is evaluated:

| Value | Description |
|-------|-------------|
| `detection` | Pattern-based matching against the PR diff. Requires at least one entry in `patterns`. |
| `policy` | Evaluates structural properties of the PR (description, commit messages, test coverage). |

```yaml
kind: detection
```

### `action`

**Type:** enum -- **Required**

The default behavior when the pack produces findings. Users can override this per-pack in the dashboard.

| Value | Description |
|-------|-------------|
| `flag` | Add findings to the risk score and include them in the reviewer brief. No inline comments. |
| `warn` | Post inline comments on the PR for each finding. Findings also affect the risk score. |
| `block` | Set the Pullminder status check to "failure", preventing the PR from being merged until findings are resolved. |

```yaml
action: flag
```

### `version`

**Type:** integer -- **Required**

Integer version of the pack. Increment each time you modify the pack's patterns or configuration.

```yaml
version: 3
```

### `schema_version`

**Type:** integer -- **Optional**

The version of the pack schema this file conforms to. Can be omitted; when present the only supported value is `1`.

```yaml
schema_version: 1
```

### `author`

**Type:** string -- **Optional**

GitHub handle of the pack author. Used for attribution in the registry and verified during publishing. Required only when publishing to the community registry.

```yaml
author: your-github-handle
```

### `max_weight`

**Type:** integer -- **Optional** -- Default: `10`

The maximum weight that any single finding from this pack can contribute to the risk score. This cap prevents a single pack from dominating the overall score.

```yaml
max_weight: 10
```

## `scoring` array

Each entry defines a scoring tier. The pack's contribution to the risk score is the highest tier whose `min_findings` threshold is met by the number of findings in the PR.

### `scoring[].min_findings`

**Type:** integer -- **Required**

Minimum number of findings from this pack required to reach this score tier.

### `scoring[].score`

**Type:** integer -- **Required**

The risk score contribution when this tier is reached.

```yaml
scoring:
  - min_findings: 1
    score: 5
  - min_findings: 3
    score: 10
  - min_findings: 5
    score: 15
```

## `patterns` array

An array of pattern objects. Required for `detection` packs. Each pattern defines a single detection rule.

### `patterns[].name`

**Type:** string -- **Required**

Human-readable name for the pattern. Displayed in the reviewer brief and dashboard findings list.

```yaml
- name: Hardcoded AWS access key
```

### `patterns[].rule_id`

**Type:** string -- **Required**

Unique identifier for the pattern within the pack. Convention is an uppercase prefix followed by a number (e.g., `SEC-001`, `GO-003`). Rule IDs must be unique across all patterns in the pack.

```yaml
  rule_id: SEC-001
```

### `patterns[].regex`

**Type:** string -- **Required**

A regular expression matched against each added or modified line in the PR diff. Uses RE2 syntax (Go-compatible). The regex is applied per-line; multiline matching is not supported.

```yaml
  regex: "AKIA[0-9A-Z]{16}"
```

### `patterns[].language`

**Type:** string -- **Required**

Restricts the pattern to files of a specific language. Use `*` to match all files. Supported values:

| Value | File extensions |
|-------|----------------|
| `*` | All files |
| `go` | `.go` |
| `python` | `.py` |
| `javascript` | `.js`, `.jsx`, `.mjs` |
| `typescript` | `.ts`, `.tsx` |
| `rust` | `.rs` |
| `ruby` | `.rb`, `.erb` |
| `php` | `.php` |
| `java` | `.java` |
| `c` | `.c`, `.h` |
| `cpp` | `.cpp`, `.cc`, `.cxx`, `.hpp` |
| `csharp` | `.cs` |
| `swift` | `.swift` |
| `kotlin` | `.kt`, `.kts` |
| `yaml` | `.yaml`, `.yml` |
| `json` | `.json` |
| `dockerfile` | `Dockerfile`, `*.dockerfile` |
| `terraform` | `.tf` |
| `shell` | `.sh`, `.bash`, `.zsh` |

```yaml
  language: go
```

### `patterns[].severity`

**Type:** enum -- **Required**

The severity level of findings produced by this pattern. Severity determines the finding's base weight in the risk score.

| Severity | Weight | Use when |
|----------|--------|----------|
| `critical` | 10 | The finding represents an immediate, exploitable security risk (e.g., leaked production credentials). |
| `error` | 8 | Serious error that should be fixed before merging (e.g., SQL injection, command injection). |
| `high` | 7 | The finding is a serious issue that should be resolved before merging (e.g., unvalidated input in a sensitive path). |
| `medium` | 5 | The finding is a notable concern that warrants reviewer attention (e.g., missing input validation). |
| `low` | 3 | The finding is a minor issue or style violation (e.g., debug logging left in production code). |
| `info` | 1 | The finding is informational and does not significantly affect the risk score (e.g., a TODO comment). |

```yaml
  severity: high
```

### `patterns[].category`

**Type:** string -- **Required**

Freeform category used for grouping and filtering findings in the dashboard. Common values include `security`, `code-quality`, `testing`, `infrastructure`, and `dependencies`.

```yaml
  category: security
```

### `patterns[].description`

**Type:** string -- **Optional**

A longer explanation of what the pattern detects and why it matters. Displayed in the reviewer brief and finding detail views.

```yaml
  description: >
    AWS access keys should never appear in source code.
    Use environment variables or a secrets manager instead.
```

### `patterns[].fix_templates`

**Type:** array of strings -- **Optional**

Suggested fixes displayed alongside the finding. Each string is a separate suggestion. Providing fix templates helps developers resolve findings quickly.

```yaml
  fix_templates:
    - "Store the key in AWS Secrets Manager and reference it via environment variable."
    - "Use IAM roles instead of static access keys."
```

## `overrides` object

Exclusion rules that apply to all patterns in the pack.

### `overrides.ignore_paths`

**Type:** array of strings -- **Optional**

Glob patterns for file paths that should be excluded from pattern matching. Useful for skipping test fixtures, vendored code, or generated files.

```yaml
overrides:
  ignore_paths:
    - "**/vendor/**"
    - "**/testdata/**"
    - "**/*.generated.go"
```

### `overrides.ignore_authors`

**Type:** array of strings -- **Optional**

GitHub usernames whose PRs should be excluded from this pack's evaluation. Useful for skipping automated accounts.

```yaml
overrides:
  ignore_authors:
    - "dependabot[bot]"
    - "renovate[bot]"
```

## Schema version history

| Version | Changes |
|---------|---------|
| `1` | Initial schema. Supports detection and policy pack kinds, additive scoring model, pattern-based matching, and path/author overrides. |

Future schema versions will be backward-compatible where possible. Packs specifying an older `schema_version` will continue to work with newer versions of Pullminder.

## Complete example

```yaml
slug: node-security
name: Node.js Security
kind: detection
action: warn
version: 3
max_weight: 10

scoring:
  - min_findings: 1
    score: 5
  - min_findings: 3
    score: 10
  - min_findings: 5
    score: 15

patterns:
  - name: Dynamic code execution
    rule_id: NODE-001
    regex: "\\beval\\s*\\("
    language: javascript
    severity: error
    category: security
    description: >
      Dynamic code execution is a common vector for injection
      attacks. Use safer alternatives like JSON.parse() for data
      or Function() with strict input validation.
    fix_templates:
      - "Replace with JSON.parse() if parsing JSON data."
      - "Use a sandboxed execution environment if dynamic code evaluation is required."

  - name: Child process with shell option
    rule_id: NODE-002
    regex: "child_process.*shell\\s*:\\s*true"
    language: javascript
    severity: error
    category: security
    description: >
      Spawning child processes with shell: true enables shell
      interpretation of the command string, which can lead to
      command injection if any part of the string is user-controlled.
    fix_templates:
      - "Use execFile() or spawn() without the shell option and pass arguments as an array."

  - name: Unvalidated redirect
    rule_id: NODE-003
    regex: "res\\.redirect\\(\\s*req\\.(query|body|params)"
    language: javascript
    severity: medium
    category: security
    description: >
      Redirecting to a URL taken directly from user input can
      lead to open redirect vulnerabilities.
    fix_templates:
      - "Validate the redirect URL against an allowlist of permitted destinations."

overrides:
  ignore_paths:
    - "**/test/**"
    - "**/tests/**"
    - "**/__tests__/**"
  ignore_authors:
    - "dependabot[bot]"
```

## Documentation

Full documentation is available at [docs.pullminder.com](https://docs.pullminder.com/packs/overview/).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to create and submit rule packs.

## Security

To report a vulnerability, please email **security@pullminder.com**. See [SECURITY.md](https://github.com/pullminder/.github/blob/main/SECURITY.md) for the full policy.

## License

[Apache-2.0](LICENSE)

---

_This README is auto-generated from the [pullminder.com monorepo](https://github.com/upmate/pullminder.com). Last synced: 2026-04-18._
