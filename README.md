# NeetoPlanner CLI

A command-line interface for NeetoPlanner. Manage projects, lists and todos, and set per-workspace defaults.

> **Status:** `login`, `logout`, `whoami`, `doctor`, `version`, `commands`,
> `completion` and `setup` work today. `projects list`, `lists list`, `lists show`,
> `todos list`, `todos create`, `todos update` and project defaults are wired to
> the API. `todos show` and `todos done` still exit with a message naming the
> endpoint they are waiting on. Track
> [neeto-planner-web#12900](https://github.com/neetozone/neeto-planner-web/issues/12900).

## Command reference

```
neetoplanner projects list
neetoplanner lists list [--project <p>]
neetoplanner lists show <sid> [--project <p>]
neetoplanner todos list [--project <p>]
neetoplanner todos show <id>
neetoplanner todos create "Ship it" [--project <p>] [--list <l>]
neetoplanner todos update <id> [--title ...]
neetoplanner todos done <id>
neetoplanner config set default-project <p>
```

`--project` accepts a project SID. Commands resolve it in this order:
the `--project` flag, then `NEETOPLANNER_PROJECT`, then the default saved per
subdomain by `neetoplanner config set default-project`.

### Create a promotion todo

Find the project and list SIDs using `projects list` and `lists list`, then run:

```bash
neetoplanner todos create "Promote the NeetoCal changelog" \
  --project <project-sid> --list <list-sid> \
  --description "Post on X, LinkedIn and the Neeto community. https://app.neetocal.com/changelog" \
  --idempotency-key "engage:<workspace-id>:<post-id>:published" --json
```

`--list` takes a list SID and is optional; omitting it creates a todo without a
list. `--description` is optional. `--json` returns the todo under `data`,
including its UUID `id`, numeric `handle`, and `url`. `--quiet` prints only the UUID.
Use the numeric handle with `todos update`.

Reuse the same `--idempotency-key` when retrying one event. Concurrent requests
with the same key and project return the same todo without overwriting it.
Keys are limited to 255 characters. Without a key, each call creates a new todo.
Permanently deleting or moving the todo removes deduplication from that project;
clones do not inherit keys. Assignee and due-date options are not supported by
`todos create`.

This command requires Planner's external todo creation API. Deploy the API and
its migration before releasing this CLI version.

<!-- neeto-cli-commons:installation:start -->
## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew install neetozone/tap/neetoplanner
```

**Shell script:**

```bash
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoPlanner/latest/install.sh | sh
```

This verifies the download's SHA-256 checksum against the published `SHA256SUMS`,
then installs to `/usr/local/bin` (may prompt for sudo). Set `NEETOPLANNER_INSTALL_DIR`
to a directory you own to install without sudo.

### Windows

**PowerShell:**

```powershell
irm https://neeto-downloads.s3.amazonaws.com/cli/NeetoPlanner/latest/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoPlanner/latest/install.cmd -o install.cmd && install.cmd
```

Both verify the download's SHA-256 checksum before installing to
`%LOCALAPPDATA%\Programs\neetoplanner` and adding it to your user PATH. Set
`NEETOPLANNER_INSTALL_DIR` to install somewhere else.
<!-- neeto-cli-commons:installation:end -->

<!-- neeto-cli-commons:verify-installation:start -->
### Verify installation

```bash
neetoplanner --help
```
<!-- neeto-cli-commons:verify-installation:end -->

<!-- neeto-cli-commons:ai-coding-assistants:start -->
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
run these commands from the root of the project the assistant works in. Re-run
them after every upgrade: `setup cursor` and `setup windsurf` overwrite their rule
file, while `setup copilot`, `setup gemini` and `setup codex` keep the existing
content of their file and replace only the NeetoPlanner section instead of adding a
duplicate.
<!-- neeto-cli-commons:ai-coding-assistants:end -->

<!-- neeto-cli-commons:prerequisites:start -->
## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoPlanner organization
<!-- neeto-cli-commons:prerequisites:end -->

## Development

```bash
git clone https://github.com/neetozone/neeto-planner-cli.git
cd neeto-planner-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

<!-- neeto-cli-commons:make-targets:start -->
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
<!-- neeto-cli-commons:make-targets:end -->

### Pointing to a local or staging server

Set `NEETOPLANNER_BASE_URL` to override the default `https://<subdomain>.neetoplanner.com`:

```bash
export NEETOPLANNER_BASE_URL=http://spinkart.lvh.me:8830
neetoplanner login --subdomain spinkart
```

<!-- neeto-cli-commons:global-flags:start -->
## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |
| `--verbose` | Expand every field of a record instead of a table. |
<!-- neeto-cli-commons:global-flags:end -->

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

<!-- neeto-cli-commons:release:start -->
## Release

Releases are cut by the CI pipeline defined in `.neetoci/release.yml`.

Merging a PR with a `major`, `minor`, or `patch` label to `main` triggers the shared release script from `neeto-cli-commons`. The script:

* Bumps and tags `VERSION`
* Runs GoReleaser
* Uploads artifacts to `s3://neeto-downloads/cli/NeetoPlanner/`
* Updates the Homebrew tap (`neetozone/tap`)
* Pushes the version bump commit to `main`
<!-- neeto-cli-commons:release:end -->
