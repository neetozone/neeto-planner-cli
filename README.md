# NeetoPlanner CLI

A command-line interface for NeetoPlanner.

> **Status:** `login`, `logout`, `whoami`, `doctor`, `version`, `commands`,
> `completion` and `setup` work today. The `projects`, `lists` and `todos`
> commands are registered with their final flags and help text, but each one
> exits with a message naming the endpoint it is waiting on — the NeetoPlanner
> external API is still being built. Track
> [neeto-planner-web#12675](https://github.com/neetozone/neeto-planner-web/issues/12675).

## Command reference

```
neetoplanner projects list
neetoplanner lists list [--project <p>]
neetoplanner todos list [--project <p>] [--list <l>]
neetoplanner todos show <id>
neetoplanner todos create "Ship it" [--project <p>] [--list <l>]
neetoplanner todos update <id> [--title ...]
neetoplanner todos done <id>
neetoplanner config set default-project <p>
```

`--project` accepts a project name or ID. Commands resolve it in this order:
the `--project` flag, then `NEETOPLANNER_PROJECT`, then the default saved per
subdomain by `neetoplanner config set default-project`.

## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew install neetozone/homebrew-tap/neetoplanner
```

**Shell script:**

```bash
curl -fsSL https://neetoplanner.com/cli/install.sh | sh
```

### Windows

**PowerShell:**

```powershell
irm https://neetoplanner.com/cli/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neetoplanner.com/cli/install.cmd -o install.cmd && install.cmd
```

### Verify installation

```bash
neetoplanner --help
```

## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoPlanner organization

## Development

```bash
git clone https://github.com/neetozone/neeto-planner-cli.git
cd neeto-planner-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

### Make targets

```bash
make build          # Builds ./neetoplanner
make test           # Run tests
make lint           # golangci-lint
make fmt            # gofmt -w
make vet            # go vet
make check          # fmt + vet + test
make install        # Installs to /usr/local/bin
make clean          # Remove built binary
```

### Pointing to a local or staging server

Set `NEETOPLANNER_BASE_URL` to override the default `https://<subdomain>.neetoplanner.com`:

```bash
export NEETOPLANNER_BASE_URL=http://acme.lvh.me:8980
neetoplanner login --subdomain acme
```

## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

## Release

Releases are cut by BigBinary's CI pipeline defined in
`.neetoci/release.yml`. Merging a PR with a `major` / `minor` / `patch`
label to `main` triggers `.scripts/release.sh`, which tags the current
VERSION, runs GoReleaser, uploads artifacts to
`s3://neeto-downloads/cli/NeetoPlanner/`, updates the Homebrew tap
(`neetozone/homebrew-tap`), and opens the next-version bump PR.

## AI coding assistants

```bash
neetoplanner setup claude      # Register plugin with Claude Code
neetoplanner setup cursor      # Write .cursor/rules/neetoplanner.mdc
neetoplanner setup windsurf    # Write .windsurf/rules/neetoplanner.md
neetoplanner setup copilot     # Add a NeetoPlanner section to .github/copilot-instructions.md
neetoplanner setup gemini      # Add a NeetoPlanner section to GEMINI.md
neetoplanner setup codex       # Add a NeetoPlanner section to AGENTS.md
```

Every command except `setup claude` writes into the current project directory, so
run these commands from the root of the project the assistant works in.
`setup copilot`, `setup gemini` and `setup codex` keep the existing content of
their file and, when re-run after an upgrade, replace the NeetoPlanner section instead
of adding a duplicate.
