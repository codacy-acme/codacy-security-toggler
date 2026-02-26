# codacy-security-toggler

A CLI tool that bulk-enables or bulk-disables **Security-category code patterns** across an entire Codacy organisation — covering both repositories that follow a coding standard and those that are detached from one.

## How it works

The tool runs two phases in sequence:

### Phase 1 — Coding standards

1. Fetches all coding standards for the organisation (or a single one by ID).
2. For each standard that is **not a draft**, creates a new draft from it using the same name and languages (`sourceCodingStandard` parameter).
3. Lists every tool configured in the draft.
4. Bulk-updates all Security-category patterns for each tool (`categories=Security`).
5. Optionally promotes the draft to an effective coding standard.

### Phase 2 — Detached repositories

1. Lists all organisation repositories with analysis data.
2. Filters to those whose `standards` field is empty (not following any coding standard).
3. For each detached repository, lists its analysis tools.
4. Bulk-updates all Security-category patterns for each tool directly on the repository.

## Requirements

- Go 1.22+
- A Codacy **account** API token — generate one at **Account → API tokens** in the Codacy UI. Repository tokens are not accepted.

## Build

```bash
go build -o codacy-security-toggler .
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--api-token` | — | Codacy API token. Can also be set via `CODACY_API_TOKEN`. |
| `--provider` | `gh` | Git provider: `gh` (GitHub), `gl` (GitLab), `bb` (Bitbucket). |
| `--organization` | — | Organisation name on the Git provider **(required)**. |
| `--coding-standard-id` | `0` | ID of a specific coding standard to process. `0` processes all standards. |
| `--enable` | `true` | `true` to enable security patterns, `false` to disable them. |
| `--promote` | `true` | Promote the updated draft to an effective coding standard. |
| `--skip-live` | `false` | Skip standards that are not drafts instead of creating a new draft from them. |
| `--dry-run` | `false` | Print what would happen without making any API changes. |
| `--verbose` | `false` | Print additional detail such as tool names and UUIDs. |

## Examples

### Enable security patterns across the whole organisation

```bash
./codacy-security-toggler \
  --api-token="$CODACY_API_TOKEN" \
  --provider=gh \
  --organization=my-org \
  --enable=true
```

### Disable security patterns on a specific coding standard

```bash
./codacy-security-toggler \
  --api-token="$CODACY_API_TOKEN" \
  --organization=my-org \
  --coding-standard-id=42 \
  --enable=false
```

### Dry run before making changes

```bash
./codacy-security-toggler \
  --api-token="$CODACY_API_TOKEN" \
  --organization=my-org \
  --enable=true \
  --dry-run=true
```

### Enable patterns but keep the result as a draft (skip promote)

```bash
./codacy-security-toggler \
  --api-token="$CODACY_API_TOKEN" \
  --organization=my-org \
  --enable=true \
  --promote=false \
  --verbose=true
```

## Authentication

Pass the token via the `--api-token` flag or export it as an environment variable:

```bash
export CODACY_API_TOKEN=your_token_here
```

Tokens must be account-level API tokens as required by Codacy API v3. Repository tokens are not accepted.
