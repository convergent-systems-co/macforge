Build a macOS setup CLI tool in Go called `macheim` (Old Norse for "Mac home") that
bootstraps and continuously syncs a Mac to a known-good state defined in a git repo.

CLI FRAMEWORK
Use urfave/cli/v3 (github.com/urfave/cli/v3). Subcommand structure, flag inheritance,
before/after hooks, and shell completion are all supported — no need for cobra.

CORE CONCEPT
- The git repo IS the source of truth. The binary is a thin client.
- Config files (Brewfile, dotfiles, macOS defaults, etc.) live in the repo as plain files.
- They're also `//go:embed`-ed at build time as a fallback so a fresh Mac with no repo
  cloned can still bootstrap from just the binary.
- Once the repo is cloned, the binary prefers repo files over embedded ones.

REPO DISCOVERY (in priority order)
1. `--repo` flag
2. `MACHEIM_REPO` env var
3. `~/.config/macheim/config.yaml` → `repo_path:`
4. Convention: `~/src/macheim` then `~/code/macheim`
5. Fall back to embedded files (read-only mode; `update local-to-remote` will error
   with a clear message telling the user to clone the repo first)

COMMANDS

  macheim bootstrap
      Run everything end-to-end on a fresh Mac:
      brew install → brew bundle → zsh setup → dotfiles apply → macos defaults → downloads

  macheim brew install
      Install Homebrew itself. Detect /opt/homebrew/bin/brew (Apple Silicon) and
      /usr/local/bin/brew (Intel). Skip if present. Stream the official install script
      output live (embed it via //go:embed; see notes). Append `eval "$(brew shellenv)"`
      to the right rc file idempotently after install. Verify via `brew --version`.

  macheim brew bundle
      Run `brew bundle --file=<repo>/Brewfile`. Skipped formulae print a friendly note.

  macheim zsh setup        (stub for now)
  macheim dotfiles apply   (stub for now)
  macheim macos defaults   (stub for now)
  macheim downloads        (stub for now)

  macheim update local-to-remote
      Review what's drifted on this Mac vs. the repo, and update the REPO files to
      match local state. For each module:
        brew:        diff `brew bundle dump --describe --file=-` against repo Brewfile.
                     Show added/removed formulae, casks, taps, mas apps. Prompt to apply.
                     On confirm, rewrite <repo>/Brewfile.
        dotfiles:    diff $HOME dotfiles against <repo>/dotfiles/. Show changed/new
                     files. Prompt before copying changes back to the repo.
        defaults:    (stub) read current `defaults` values and diff against repo manifest.
      Always end with: "Repo updated. Review with `git diff`, then commit and rebuild
      with `make build` to refresh the embedded fallback."
      Flags: --module=brew|dotfiles|defaults|all (default all), --yes (skip prompts),
             --dry-run.

  macheim update remote-to-local
      Pull the repo (`git -C <repo> pull --ff-only`) then apply changes to this Mac:
        brew:        `brew bundle --file=<repo>/Brewfile`. Optionally `--cleanup` to
                     remove formulae no longer in the Brewfile (gated behind --prune).
        dotfiles:    copy <repo>/dotfiles/ → $HOME, with backups of any overwritten
                     files to ~/.macheim-backups/<timestamp>/.
        defaults:    (stub) apply repo manifest via `defaults write`.
      Flags: --module=..., --prune, --no-pull, --yes, --dry-run.

  macheim status
      Quick read-only summary: is brew installed, what's drifted in each module, last
      git commit in the repo, embedded-fallback version. No changes made.

  macheim doctor
      Sanity-check the environment: xcode-select, brew, repo path, write permissions,
      shell rc files, etc. Exit non-zero if anything's broken.

GLOBAL FLAGS (inherited)
  --repo PATH        override repo location
  --dry-run          print actions, change nothing
  --verbose, -v      extra output
  --quiet, -q        suppress non-error output
  --yes, -y          skip confirmation prompts
  --no-color         disable colored output

PROJECT LAYOUT
  macheim/
    main.go
    cmd/
      root.go                  // app := &cli.Command{...}
      bootstrap.go
      brew.go                  // brew install, brew bundle
      update.go                // update local-to-remote, update remote-to-local
      status.go
      doctor.go
    internal/
      config/
        config.go              // load ~/.config/macheim/config.yaml, repo discovery
      embedded/
        embedded.go            // //go:embed scripts/ configs/
        scripts/
          install-brew.sh
        configs/
          Brewfile             // fallback Brewfile (synced from repo at build time)
          dotfiles/            // fallback dotfiles
      shell/
        detect.go              // detect zsh/bash, find rc files
        run.go                 // exec.Cmd helper with live stdout/stderr streaming
        prompt.go              // y/n prompts, respects --yes
      brew/
        detect.go              // is brew installed? which path? arch-aware
        install.go             // install brew
        bundle.go              // brew bundle wrapper
        diff.go                // dump-vs-repo diff for local-to-remote
      dotfiles/
        diff.go
        apply.go
      gitrepo/
        repo.go                // pull, diff, status helpers (use os/exec on `git`,
                               // don't pull in go-git unless needed)
    go.mod
    Makefile
    README.md

EMBED STRATEGY
- /internal/embedded/scripts/install-brew.sh is a pinned snapshot of Homebrew's installer.
  Document in a comment how to refresh it (curl the official URL, review the diff, commit).
- /internal/embedded/configs/* is a snapshot of the repo's own config files. The Makefile's
  `build` target should `cp` the latest repo files into this dir before `go build` so the
  embedded fallback stays current. Add `make embed-sync` as the explicit step.

DESIGN PRINCIPLES
- Idempotent everywhere. Re-running anything is safe and prints "no changes" when there's
  nothing to do.
- Every mutating operation respects --dry-run AND prompts unless --yes.
- Architecture-aware (arm64 vs amd64) for brew paths.
- Errors are actionable: include the exact command the user can run to recover.
- No silent network calls. If a step needs the internet, say so before doing it.

DELIVERABLES
Complete, runnable code for every file. Comments explaining:
1. The //go:embed directives and the build-time sync from repo to embedded/
2. Repo-discovery priority and the read-only-fallback mode
3. The local-to-remote diff/write flow for brew (the trickiest piece)
4. The shell detection and rc-file idempotency check

Include:
- Makefile with `build`, `embed-sync`, `install` (→ /usr/local/bin), `test`, `lint`, `clean`
- README covering: install, first-time bootstrap, the two update directions with examples,
  config file format, and the "rebuild after local-to-remote" workflow
- A sample ~/.config/macheim/config.yaml
- A sample Brewfile in the repo root
