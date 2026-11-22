# nixops

**nixops** is a Linux CLI/TUI application for safely managing and cleaning a
local Nix store on systems where Nix is installed independently of the host
OS — for example Fedora, Fedora Atomic or Bazzite with standalone Nix, Home
Manager and devenv.

The compiled executable is called **`nixops`**.

> **Safety model**
>
> nixops does not delete files directly from `/nix/store`.
> It delegates store lifecycle management to Nix.
>
> nixops never runs commands through a shell (`sh -c`), never invokes
> `sudo` automatically, and never invents Nix behaviour. Every destructive
> operation supports `--dry-run`.

---

## Why nixops exists

On non-NixOS systems it is easy to accumulate large amounts of dead store
paths, obsolete devenv GC roots and stale Home Manager generations. Nix
already provides precise, safe tooling for reclaiming this space — GC roots
determine reachability, and the Nix garbage collector only removes
unreachable paths. nixops orchestrates those official tools in one place,
with previews, confirmations and a dry-run mode, instead of asking you to
remember the low-level incantations.

---

## Supported environments

- Fedora / Fedora Atomic / Bazzite (and other Linux x86_64 hosts)
- Standalone Nix installation (single- or multi-user/daemon)
- Home Manager (standalone)
- devenv
- bash / zsh / fish (nixops does not depend on your shell)

nixops is **not** a NixOS tool: it never touches `nixos-rebuild` or system
generations. The host OS is detected and shown for information only.

---

## Installation

```bash
go build -o nixops ./cmd/nixops
sudo install -m755 nixops /usr/local/bin/nixops   # optional
```

Or simply run it from the repository root after building.

`make build` (or `CGO_ENABLED=0 go build -o nixops ./cmd/nixops`) produces a
**fully static** binary with no C dependencies, so it runs on any Linux
host regardless of its C library — no dynamic interpreter or linker is
required at runtime.

---

## CLI usage

```text
nixops                       launches the interactive TUI
nixops tui                   launches the interactive TUI explicitly
nixops status                Show Nix/Home Manager/devenv status
nixops clean                 Run a configurable cleanup
nixops gc                    Garbage collect unreachable store paths
nixops gc-old                Remove old generations and garbage collect
nixops dead                  Show unreachable store paths
nixops live                  Show live store paths
nixops roots                 Show garbage collection roots
nixops optimise              Optimise/deduplicate the Nix store
nixops devenv-envs           List devenv environments under ~/Projects
nixops devenv-gc             Clean devenv GC roots for an environment
nixops devenv-update         Update an environment's inputs and lock files
nixops devenv-info           Show an environment's info (read-only)
nixops hm-generations        Show Home Manager generations
nixops hm-expire             Expire old Home Manager generations
```

Global options (available on every command):

```text
--dry-run    Show actions without performing destructive operations
--yes        Skip confirmation prompts
--verbose    Enable verbose output
--quiet      Suppress informational output
--no-color   Disable colour
--version    Show version
-h, --help   Show help
```

### Selecting a devenv environment

nixops discovers devenv projects under `~/Projects` (override with
`NIXOPS_PROJECTS_DIR`). A directory is a devenv environment when it contains
`devenv.nix`, `devenv.yaml` or `devenv.yml`.

```text
nixops devenv-envs                    list projects and their directories
nixops devenv-envs --json             machine-readable listing
nixops devenv-envs --projects-dir DIR scan a different root
```

`devenv-gc`, `devenv-update` and `devenv-info` accept:

```text
--dir, -C DIR   manage this project directory
--env NAME      select a project by its listing name
```

Without either flag, nixops uses the project containing the current
directory, or the sole discovered environment. If several exist and the
current directory is not inside one, the command fails with guidance instead
of guessing. The managed directory is always printed, and dry-run plans
include it.

### Examples

```bash
nixops status
nixops dead                        # preview what Nix considers unreachable
nixops roots                       # what is keeping paths alive
nixops gc --dry-run                # show what would be executed
nixops gc                          # normal Nix garbage collection
nixops gc-old --dry-run            # explains that old generations would go
nixops hm-expire --older-than 30d --dry-run
nixops hm-expire --older-than 30d
nixops clean --dry-run             # complete cleanup plan, nothing executed
nixops clean --yes                 # conservative cleanup, no prompts
nixops clean --home-manager --gc --optimise --yes

nixops devenv-envs                 # what can nixops manage?
nixops devenv-info --env my-app    # inspect a discovered environment
nixops devenv-update -C ~/Projects/my-app
nixops devenv-gc --dir ~/Projects/my-app --dry-run
nixops clean --dry-run --devenv-dir ~/Projects/my-app
```

### `nixops clean`

`clean` runs several operations in a safe order. Default selection
(conservative):

1. `devenv gc`
2. `home-manager expire-generations "-30 days"`
3. `nix-store --gc`

Options:

```text
--dry-run
--yes
--hm-older-than 30d
--devenv          include devenv GC
--devenv-dir DIR  devenv project for the devenv step
--devenv-env NAME devenv environment for the devenv step
--home-manager    include Home Manager generation expiry
--gc              include Nix GC
--optimise        include store optimisation
--aggressive      include aggressive GC (nix-collect-garbage -d)
```

If any of `--devenv/--home-manager/--gc/--optimise/--aggressive` is given,
only the selected steps run. Aggressive GC is never included by default.

### Exit codes

```text
0   success
1   operation failed
2   invalid command/arguments
3   required dependency missing
4   user cancelled
```

---

## TUI usage

Running `nixops` (or `nixops tui`) opens an interactive interface with the same
operations behind the scenes: status dashboard, dead-path preview, GC,
aggressive GC, optimisation, roots/live listing, Home Manager generations
and expiry, devenv environment management — plus a **Clean up** workflow
where you can select multiple operations and review the plan before running
it.

The devenv section starts with **Select environment**, which lists every
devenv project found under `~/Projects` (and the project containing the
current directory). The section header, confirmation screen, running screen
and results all show the managed directory. Choosing a devenv operation
before selecting an environment opens the picker first, then continues with
the operation. The cleanup workflow also asks for an environment when its
devenv step is selected.

The `Status` item stands alone at the top of the menu, outside the Nix
Store group. The header shows a context label that follows the screen and
menu section you are on, and `⚡ nixops` badges the app version. Dry-run
outcomes are presented uniformly for every task (a yellow `[DRY RUN]`
badge with the plan); when dry-run is enabled, destructive tasks skip
their confirmation screen since nothing will execute.

Keyboard controls:

```text
↑ / k       Move up
↓ / j       Move down
Enter       Select / expand output
Esc         Back
q           Quit
?           Help
r           Refresh
d           Toggle dry-run (shown as [DRY RUN] in the header)
y           Confirm
n           Cancel
Space       Toggle selection (cleanup workflow)
```

While an operation runs you can press `Esc` to cancel. Long-running
operations show an indeterminate spinner — nixops never fabricates
progress percentages.

---

## Dry-run mode

`--dry-run` is a first-class feature of every destructive operation. It
prints exactly what would be executed and makes no changes:

```text
DRY RUN

Garbage collection

Would execute:

  nix-store --gc

No changes have been made.
```

`--dry-run` gates **every** operation, not just destructive ones. Read-only
diagnostics such as `nixops dead`, `nixops roots`, `nixops live`,
`nixops hm-generations` and `nixops devenv-info` do not run their commands
either — they show exactly what they would run. The only commands that
still run are read-only *discovery* steps required to compute impact: for
example `nixops hm-expire --older-than 30d --dry-run` executes
`home-manager generations` (read-only) so it can show you which
generations would be affected, while refusing to execute
`home-manager expire-generations "-30 days"`. Similarly, devenv target
selection scans `~/Projects` (read-only) so dry-run output can name the
project directory; no devenv command runs. nixops never fakes results:
it will not claim a number of freed bytes it has not measured.

---

## Safety model

- Nothing is ever deleted directly from `/nix/store`.
- Store reachability and deletion decisions belong to Nix's garbage
  collector; nixops orchestrates `nix-store`, `nix-collect-garbage`,
  `home-manager` and `devenv`.
- Commands are always executed with separate arguments (never through
  `sh -c`), so user-supplied values such as `--older-than` can never be
  interpolated into a shell string.
- Destructive operations require explicit confirmation in interactive mode
  (or `--yes`); non-interactive runs without `--yes` are refused.
- devenv project discovery only stats marker files; it never modifies
  projects or devenv's GC directory. Commands run in the selected project
  directory without changing nixops's own working directory.
- The current Home Manager generation can never be selected for expiry.
- `gc-old` (aggressive GC) is always clearly separated from normal GC and
  warns that you may lose rollback points.
- No automatic `sudo`: commands run as the invoking user and Nix handles
  its own permissions. If elevated rights are needed, Nix's error is
  reported clearly.

---

## Home Manager handling

Generations are listed with `home-manager generations` and displayed in a
table. Expiry uses the official mechanism:

```bash
home-manager expire-generations "-30 days"
```

nixops converts periods like `30d`, `12w`, `6m`, `1y` into the GNU date
expression Home Manager expects, shows you exactly which generations
qualify before doing anything, and never expires the current generation.

---

## devenv handling

devenv maintains timestamped GC roots itself. nixops runs the official
`devenv gc`, which cleans obsolete roots before performing Nix garbage
collection. It never deletes anything from devenv's GC directory directly,
and `devenv-gc` is always a distinct operation from plain Nix GC.

nixops can manage any devenv project, not just the one in the current
working directory:

- `nixops devenv-envs` discovers projects under `~/Projects` (override with
  `NIXOPS_PROJECTS_DIR`), skipping hidden and generated directories
  (`node_modules`, `.devenv`, `result`, …) and limiting scan depth.
- `nixops devenv-info --env NAME` (or `--dir, -C DIR`) selects a project.
  Without a selector, nixops uses the project containing the current
  directory, or the sole discovered environment; it never guesses between
  multiple projects.
- `nixops devenv-gc` runs `devenv gc` in the selected project directory.
- `nixops devenv-update` runs `devenv update` in the selected directory,
  refreshing its inputs and lock files. It honours `--dry-run`.
- `nixops devenv-info` runs `devenv info`, a read-only overview.

The selected directory is always printed, included in dry-run plans, and
shown in the TUI. In the TUI, **devenv → Select environment** changes the
target for all three operations and the cleanup workflow. Commands are
executed with the project directory as their working directory; nixops
never changes its own process working directory or builds a shell command
string.

Discovery is read-only: it stats marker files and never modifies projects.

---

## Nix GC behaviour

- `nixops gc` runs `nix-store --gc`, which removes store paths no longer
  reachable from a GC root.
- `nixops dead` previews with `nix-store --gc --print-dead` (read-only).
- `nixops roots` shows live GC roots with `nix-store --gc --print-roots`
  and best-effort classification (Home Manager, devenv, User profiles,
  Manual roots, Other).
- `nixops live` lists reachable paths with `nix-store --gc --print-live`.
- `nixops gc-old` runs `nix-collect-garbage -d` after an explicit warning.

nixops does not reproduce the garbage collector's logic in Go; Nix decides
what is dead. Output parsing is best-effort and falls back to showing the
raw command output.

---

## Optimisation

`nixops optimise` runs `nix-store --optimise`, which deduplicates identical
files in the store by hard-linking identical content. This is separate from
garbage collection. Where before/after measurements can be taken (via
read-only `du` of the store), nixops shows Before/After/Saved; otherwise
it shows the command output. This operation can take a while on large
stores and can be cancelled in the TUI.

---

## Development

```bash
go build ./...          # compile everything
go build -o nixops ./cmd/nixops
go test ./...           # unit + integration tests
go vet ./...            # static checks
```

The test suite never touches a real Nix store:

- Unit tests use fake runners (`internal/runner.FakeRunner`) and cover
  command construction, argument validation, dry-run behaviour, Home
  Manager generation parsing, Nix GC output parsing, byte-size parsing,
  operation ordering, and devenv project discovery/selection.
- Integration tests (`tests/`) build the real binary and run it against
  fake executables in `tests/fixtures/bin/` that emit representative
  outputs, log their invocations, and optionally log their working
  directory, so destructive calls and devenv project targets can be
  asserted to have (or have not) happened.

Cross-compilation targets `linux/amd64` initially; `linux/arm64` works via
`GOARCH=arm64 go build`. Unsupported platforms are reported gracefully.

---

## License

MIT — see [LICENSE](LICENSE).

<!-- agent-sandbox:begin (managed by sandbox-note.sh, edits are overwritten) -->
## Agent sandboxes

**If the environment variable `TMAKR_POD` is `1`, you are in an agent
sandbox** (a tmakr pod); check with `printenv TMAKR_POD`. If it isn't set,
ignore this section.

In a sandbox, these instructions take precedence over anything elsewhere in
this file: devenv, direnv, `devenv shell` / `devenv up`, push and
pull-request steps.

### Git: commit, never push

- Commit your work on a branch. **Never push.** In a tmakr pod the GitHub
  token is read-only, so `git push` fails with 403 by design. Don't retry it,
  change remotes, switch to SSH, or look for other credentials.
- Tom fetches your commits from the sandbox, reviews them and pushes from his
  own machine. Uncommitted changes and stashes can't be fetched: commit before
  you say a task is done.
- Where this file says to open a PR, or to say or justify something in the
  PR, put it in your commit message body instead: Tom opens the PR from your
  commits.
- Start each piece of work on a new branch off `main`. Keep `main` matching
  `origin/main`; don't commit to it.
- Don't open pull requests or use `gh`. Once you've said a branch is done,
  don't rewrite its history (no rebase, amend or force): Tom may have fetched
  it already.

### Tools: no devenv

If this repo has a `devenv.nix`, it is for local development only.

- Don't install or bootstrap Nix, devenv or direnv. Use the tools already on
  PATH: Python 3.14 with `uv` (`uv sync`, `uv run …`; Python downloads are
  disabled), Node 24 with pnpm/bun, Go, Rust, C/C++ toolchains, ruff, eslint,
  Tailwind v4, git.
- Postgres: connection settings come from the environment. If `PGHOST` is
  unset, run `set -a; . /etc/agent-db/database.env; set +a`. Don't start
  your own database server.
- Services that `devenv up` would start aren't running. Start what you need
  in the foreground (for example `uv run manage.py runserver`), or say what
  is missing.
- If something the project needs (in `devenv.nix` or elsewhere) has no
  equivalent here, say so and stop rather than working around it.
<!-- agent-sandbox:end -->
