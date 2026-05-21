# macheim

> Old Norse for "Mac home."

A macOS setup CLI that bootstraps and continuously syncs a Mac to a
known-good state defined in a git repo. The repo is the source of truth;
the binary is a thin client that knows how to discover, install, diff,
and apply.

## Who is this for?

You set up a new Mac more than once a year, you keep tweaking your
Brewfile and dotfiles on the running machine, and you want both
directions of that drift to be reconcilable from a single command. The
target audience is engineers who already understand Homebrew, git, and
shell rc files and who would rather edit a checked-in `Brewfile` than
click through GUI installers.

It is opinionated about one thing: the repo is authoritative. Local
edits are reviewed and ratcheted back into the repo via an explicit
`update local-to-remote` step, not absorbed silently.

## Status

Foundation complete. The CLI surface (`macheim --help`), config and
repo discovery, the embedded fallback, shell and git helpers, the
`doctor` sanity check, and the read-only `status` summary all ship in
the current binary. Mutating subcommands (`bootstrap`, `brew install`,
`brew bundle`, `update local-to-remote`, `update remote-to-local`,
`dotfiles apply`, `zsh setup`, `macos defaults`, `downloads`) are
registered but currently print

```
<command>: not implemented yet (see issue #N)
```

and exit zero. The examples below describe the intended behavior so the
documentation is ready when each subcommand lands. Anywhere the doc
describes intended-not-yet-shipped behavior, the heading carries a
`(planned)` tag.

Tracker: [Epic #3](https://github.com/polliard/macheim/issues/3).

## Install

Requires Go 1.26+ and a Mac running on Apple Silicon or Intel.

```bash
# From source.
git clone https://github.com/polliard/macheim.git ~/src/macheim
cd ~/src/macheim
make build
sudo make install                 # installs to /usr/local/bin/macheim
```

Or, when you already have Go on the machine:

```bash
go install github.com/polliard/macheim@latest
```

Verify:

```bash
$ macheim --version
macheim version dev (commit abc1234, built 2026-05-20T12:00:00Z)
```

The `go install` path puts the binary into `$(go env GOBIN)` (defaults
to `~/go/bin`); add that directory to your `PATH` if it isn't already.

## Quick start: first-time bootstrap

On a freshly imaged Mac with no Homebrew installed:

```bash
# 1. Install Xcode Command Line Tools (Homebrew depends on them).
xcode-select --install

# 2. Get macheim onto the machine. Pick one:
#    a) go install github.com/polliard/macheim@latest
#    b) Download a release binary into /usr/local/bin
#    c) Clone and `make build && sudo make install`

# 3. Clone your macheim repo to the conventional location so discovery
#    finds it without flags.
git clone https://github.com/<you>/macheim.git ~/src/macheim

# 4. Drive the whole bootstrap with one command. (Planned — see Status.)
macheim bootstrap
```

`bootstrap` (planned) runs the full chain end-to-end: install Homebrew,
apply the repo `Brewfile`, set up zsh, copy dotfiles, apply macOS
defaults, fetch optional downloads. Each step is idempotent and
respects `--dry-run` and `--yes`. Until `bootstrap` lands, run the
underlying steps individually as they ship.

Confirm the result with the read-only checks:

```bash
$ macheim doctor       # exits non-zero if anything is broken
$ macheim status       # prints what's drifted between this Mac and the repo
```

## Update directions

macheim recognizes two reconciliation flows. Pick the one that matches
which side has the change you want to keep.

```
+---------+   update local-to-remote   +-----------------+
|  this   | -------------------------> |  the macheim    |
|   Mac   | <------------------------- |  repo (truth)   |
+---------+   update remote-to-local   +-----------------+
```

- `update local-to-remote` — you installed something by hand or edited
  a dotfile in place; ratchet the change back into the repo.
- `update remote-to-local` — a teammate (or past-you) committed a
  change to the repo; apply it to this Mac.

Both directions accept `--module=brew|dotfiles|defaults|all` (default
`all`), `--yes` (skip prompts), and `--dry-run` (print what would
happen, change nothing).

### `macheim update local-to-remote` (planned)

For each module the command compares the live machine to the repo and
prompts before writing the diff back into the repo.

Worked example — you `brew install`ed `fd` this morning and renamed
`~/.gitconfig` to enable a new alias:

```bash
$ macheim update local-to-remote
brew:
  + brew "fd"                         # present locally, missing from Brewfile
dotfiles:
  ~ dotfiles/.gitconfig               # local content differs from repo

Apply these changes to the repo? [y/N] y
  wrote /Users/you/src/macheim/Brewfile
  wrote /Users/you/src/macheim/dotfiles/.gitconfig

Repo updated. Review with `git diff`, then commit and rebuild with
`make build` to refresh the embedded fallback.
```

The final two sentences are the load-bearing detail. Read the
[Rebuild after local-to-remote](#rebuild-after-local-to-remote) section
before declaring the change complete.

### `macheim update remote-to-local` (planned)

For each module the command pulls the repo, then applies the repo's
current state to this Mac.

Worked example — a teammate added `ripgrep` to the Brewfile and updated
`dotfiles/.zshrc`:

```bash
$ macheim update remote-to-local
git: pulled 3 commits from origin/main (ff-only)
brew: brew bundle --file=/Users/you/src/macheim/Brewfile
  installing ripgrep...
dotfiles:
  ~ ~/.zshrc                          # backed up to ~/.macheim-backups/2026-05-20T15-32-11Z/.zshrc
  copied dotfiles/.zshrc -> ~/.zshrc

Apply complete.
```

Pulling is on by default; `--no-pull` skips the `git pull` step and
applies whatever is already checked out locally.

#### `--prune` (planned)

The brew step does not uninstall anything unless `--prune` is set:

```bash
$ macheim update remote-to-local --prune
brew: brew bundle --file=Brewfile --cleanup
  uninstalling htop (not in Brewfile)
```

Warning: `--prune` runs `brew bundle cleanup`, which uninstalls every
formula and cask that is not listed in the repo Brewfile. Run with
`--dry-run` first if you are unsure what is on the chopping block.

## Config file format

Location: `~/.config/macheim/config.yaml`. Optional. When absent,
macheim continues through the discovery chain below.

Canonical starter: [`examples/config.yaml`](examples/config.yaml). Copy
it into place:

```bash
mkdir -p ~/.config/macheim
cp examples/config.yaml ~/.config/macheim/config.yaml
```

Schema:

| Field       | Type   | Default | Purpose                                                      |
|-------------|--------|---------|--------------------------------------------------------------|
| `repo_path` | string | unset   | Absolute path to your macheim repo clone. Resolved before convention paths but after `--repo` and `MACHEIM_REPO`. |

Example:

```yaml
repo_path: /Users/you/src/macheim
```

Unset fields fall through to the next step in the discovery chain.

## Repo discovery

macheim resolves the repo path on every invocation using this chain, in
order. The first match wins:

1. `--repo PATH` flag (highest priority).
2. `MACHEIM_REPO` environment variable.
3. `repo_path:` in `~/.config/macheim/config.yaml`.
4. Convention: `~/src/macheim` if it exists as a directory, then
   `~/code/macheim`.
5. **Embedded fallback** — no repo on disk. macheim runs in read-only
   mode using files baked into the binary at build time. `update
   local-to-remote` refuses to mutate in this mode and prints a message
   directing you to clone the repo first.

`macheim doctor` and `macheim status` both print which step of the
chain resolved (`flag`, `env`, `config`, `convention:src`,
`convention:code`, or unset for embed-fallback) so you can debug a
surprise.

The embedded fallback is what makes a fresh-Mac bootstrap possible
before the repo is cloned — `macheim bootstrap` works with nothing but
the binary. See [the rebuild section](#rebuild-after-local-to-remote)
for the catch.

## Rebuild after local-to-remote

This is the easiest detail to miss and the one that turns "I changed my
Brewfile" into "I changed my Brewfile and the embedded fallback is now
stale."

### The model

```
              edit                ratchet                rebuild
running Mac  ------>  running Mac  ----->  repo  ------>  new binary
  (live)              (live, diff)         (truth)        (embedded
                                                          fallback
                                                          refreshed)
```

The embedded fallback is a static snapshot baked into the binary at
`go build` time by `make embed-sync`. The sequence:

1. `update local-to-remote` writes new files into the repo working
   tree.
2. You commit and push from the repo.
3. The repo now reflects the new truth, but the binary on your `PATH`
   was built before step 1 — its embedded fallback is the old snapshot.
4. `make build` re-runs `embed-sync`, copying the freshly-committed
   repo files into `internal/embedded/configs/`, and produces a new
   binary whose fallback matches the repo.
5. `sudo make install` (or `go install`) replaces the on-`PATH` copy.

If you stop after step 2, this Mac is fine because it reads the repo
directly. The risk is on the *next* fresh Mac that bootstraps from the
binary alone with no repo cloned — it will use the stale embedded
fallback. The CLI says this back to you on every successful
`update local-to-remote`:

> Repo updated. Review with `git diff`, then commit and rebuild with
> `make build` to refresh the embedded fallback.

### FAQ: why do I need to rebuild?

The repo is the source of truth at runtime *when the repo is
reachable*. On a fresh Mac with nothing but the binary, the only files
macheim can read are the ones embedded at build time. `make build`
copies the repo into the binary so the binary's "no-repo" mode stays
current.

### FAQ: can I skip the rebuild if I never bootstrap fresh Macs?

Yes. The rebuild is only required for fresh-Mac bootstrap. If every
Mac you run macheim on has the repo cloned, you can ignore the
rebuild step indefinitely.

## Doctor and status

Both commands are read-only and shipped today.

```bash
$ macheim doctor
xcode-select:            ok      /Library/Developer/CommandLineTools
homebrew:                ok      /opt/homebrew/bin/brew
repo:                    ok      /Users/you/src/macheim (convention:src)
shell rc:                ok      ~/.zshrc writable

$ macheim status
brew:        ok          12 formulae, 3 casks installed
repo:        ok          last commit 2026-05-19 14:22 (HEAD = a1b2c3d)
drift:       unknown     (placeholder until #14 lands)
```

Run `doctor` after install and any time something feels off. Run
`status` before and after an `update` to confirm what changed.

## Commands reference

| Command                          | Status     | Purpose                                                       |
|----------------------------------|------------|---------------------------------------------------------------|
| `macheim bootstrap`              | planned    | End-to-end fresh-Mac setup. See [Quick start](#quick-start-first-time-bootstrap). |
| `macheim brew install`           | planned    | Install Homebrew itself, arch-aware.                          |
| `macheim brew bundle`            | planned    | `brew bundle --file=<repo>/Brewfile`.                         |
| `macheim zsh setup`              | planned    | Configure zsh as the macheim-managed shell.                   |
| `macheim dotfiles apply`         | planned    | Copy `<repo>/dotfiles/` into `$HOME`, with backups.           |
| `macheim macos defaults`         | planned    | Apply the repo macOS defaults manifest.                       |
| `macheim downloads`              | planned    | Fetch optional downloads listed in the repo.                  |
| `macheim update local-to-remote` | planned    | Ratchet local drift back into the repo. See [Update directions](#update-directions). |
| `macheim update remote-to-local` | planned    | Apply repo state to this Mac. See [Update directions](#update-directions). |
| `macheim status`                 | shipped    | Read-only summary of drift. See [Doctor and status](#doctor-and-status). |
| `macheim doctor`                 | shipped    | Sanity-check the environment.                                 |

Global flags (inherited by every subcommand):

| Flag                | Purpose                                                            |
|---------------------|--------------------------------------------------------------------|
| `--repo PATH`       | Override repo discovery. Also reads `MACHEIM_REPO`.                |
| `--dry-run`         | Print actions, change nothing.                                     |
| `--verbose`, `-v`   | Extra output.                                                      |
| `--quiet`, `-q`     | Suppress non-error output. Mutually exclusive with `--verbose`.    |
| `--yes`, `-y`       | Skip confirmation prompts.                                         |
| `--no-color`        | Disable colored output.                                            |

## Troubleshooting

- **`macheim --version` prints `dev (commit unknown, ...)`** — built
  outside a git checkout. Run `make build` from the cloned repo to get
  real version metadata.
- **`brew bundle` complains about a missing tap** — taps land before
  formulae; the starter `Brewfile` taps `homebrew/cask-fonts` first.
  If you added a tap by hand, run
  `macheim update local-to-remote --module=brew` to write it into the
  repo Brewfile.
- **`update local-to-remote` says no repo configured** — you are in
  embed-fallback mode. Clone the repo to `~/src/macheim` (or set
  `--repo` / `MACHEIM_REPO`) and try again.
- **`make build` reports a dirty `internal/embedded/configs/Brewfile`
  diff** — this is `embed-sync` doing its job: it copies the
  repo-root `Brewfile` into the embedded tree before each build. Run
  `make build` and commit the embedded copy along with the source.
- **`mas` lines in the Brewfile are skipped** — install the `mas` CLI
  with `brew install mas` first; `brew bundle` skips Mac App Store
  entries when `mas` is not on `PATH`.

## Contributing

Design specs live under [`docs/superpowers/specs/`](docs/superpowers/specs/).
Read [`GOALS.md`](GOALS.md) for the high-level vision before opening a
PR. CI runs `golangci-lint` and `go test -race ./...` on every push.

## Verifying the embed-sync round-trip

Sub-issue [#93](https://github.com/polliard/macheim/issues/93) calls
for a check that the embedded fallback (`internal/embedded/configs/`)
matches the repo's own source files after `make embed-sync` runs.
Until a `make check-embed` target lands, run the check manually before
cutting a release:

```bash
# Run embed-sync and confirm the embedded Brewfile matches the source.
make embed-sync && diff Brewfile internal/embedded/configs/Brewfile
# Expect: no output (the two files are byte-identical).

# Idempotency: a second invocation against an unchanged tree changes
# nothing on disk and the diff still passes.
make embed-sync && diff Brewfile internal/embedded/configs/Brewfile
# Expect: no output.
```

If the diff produces output, `embed-sync` did not run or did not
complete; rerun `make build`, which depends on `embed-sync` and will
report any underlying error. The same procedure applies to
`dotfiles/` once that tree exists at the repo root.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
