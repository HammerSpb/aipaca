# 🦙 aipaca

**aipaca** - the adorable AI config manager that herds your configurations across repositories.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What is aipaca?

**aipaca** helps you manage AI-related configuration files (Claude, Cursor, Copilot, etc.) across multiple repositories. Think of it as "dotfiles management" specifically designed for AI tool configurations, but with the charm of an alpaca. 🦙

No more copy-pasting configs between repos. Let aipaca do the heavy lifting!

## The Problem

Modern development involves multiple AI tools, each with their own configuration:
- `.claude/` - Claude Code agents and commands
- `.cursor/` - Cursor IDE rules and MCP settings
- `CLAUDE.md` - Project instructions for Claude
- Various other AI-related configs

Managing these across multiple repositories is tedious:
- You want consistent AI setups across projects
- You need different configurations for different contexts
- You want to share code without exposing AI configurations
- You need to experiment with new prompts without losing working ones

## The Solution

**aipaca** provides a "checkout/commit" style workflow for AI configurations:

```
┌─────────────────────────────────────────────────────────────────┐
│                      PROFILE STORAGE                             │
│  ~/.aipaca/profiles/                                             │
│    ├── default/        (standard AI setup)                       │
│    ├── minimal/        (lightweight config)                      │
│    ├── experimental/   (testing new prompts)                     │
│    └── project-foo/    (project-specific)                        │
└─────────────────────────────────────────────────────────────────┘
                            │
                            │ aipaca apply <profile>
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                        YOUR REPO                                 │
│    ├── .claude/                                                  │
│    ├── .cursor/                                                  │
│    └── CLAUDE.md                                                 │
│                                                                  │
│    (work, modify, experiment...)                                 │
└─────────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
            ▼               ▼               ▼
     aipaca save     aipaca save     aipaca restore
    (update profile)  --as new-name   (get originals back)
```

## Features

- 🦙 **Profile Management**: Store multiple AI configurations as named profiles
- 🚀 **Apply Profiles**: Deploy any profile to any repository with one command
- 💾 **Save Changes**: Save modifications back to profiles or create new ones
- 🔄 **Automatic Backups**: Every operation backs up existing files before changes
- ⏪ **Restore**: Instantly restore original files from backup
- 🧹 **Clean Mode**: Remove AI files for clean commits/PRs
- 📊 **Diff**: See what changed between repo and profile
- 👀 **Dry Run**: Preview any operation before executing

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/HammerSpb/aipaca.git
cd aipaca

# Build
go build -o aipaca ./cmd/aipaca

# Install (choose one)
# Option 1: Copy to PATH
sudo cp aipaca /usr/local/bin/

# Option 2: Add to PATH
echo 'export PATH="$HOME/path/to/aipaca:$PATH"' >> ~/.zshrc

# Option 3: Go install
go install ./cmd/aipaca
```

### Using Go

```bash
go install github.com/HammerSpb/aipaca/cmd/aipaca@latest
```

## Quick Start

### 1. Initialize

```bash
# Initialize aipaca (creates config and storage directory)
aipaca init

# Or initialize and import current repo's AI files as "default" profile
aipaca init --import
```

### 2. Check Status

```bash
aipaca status
```

Output:
```
Repo: /Users/you/your-project

No profile currently applied

AI files in repo:
  .claude/ (5 files)
  .cursor/ (3 files)
  CLAUDE.md
```

### 3. Save Current Setup as a Profile

```bash
# Save current AI files as a new profile
aipaca save --as my-setup
```

### 4. Apply a Profile

```bash
# Apply a profile to current directory
aipaca apply my-setup

# Preview what would happen
aipaca apply my-setup --dry-run
```

## Commands

### `aipaca init`

Initialize aipaca configuration and storage.

```bash
# Basic initialization
aipaca init

# Initialize and import current repo's AI files as "default" profile
aipaca init --import
```

### `aipaca apply <profile> [repo-path]`

Apply a profile to a repository.

```bash
# Apply to current directory
aipaca apply default

# Apply to specific repo
aipaca apply default /path/to/repo

# Preview changes without applying
aipaca apply default --dry-run

# Apply without creating backup (dangerous!)
aipaca apply default --no-backup
```

**What it does:**
1. Backs up existing AI files in the repo
2. Removes existing AI files
3. Copies profile files to the repo
4. Records the state for future operations

### `aipaca save [profile] [repo-path]`

Save repository AI files to a profile.

```bash
# Update the currently applied profile
aipaca save

# Update a specific profile
aipaca save default

# Create a new profile
aipaca save --as my-new-profile

# Overwrite existing profile
aipaca save --as existing-profile --force

# Preview what would be saved
aipaca save --dry-run
```

### `aipaca restore [repo-path]`

Restore original AI files from backup.

```bash
# Restore from the automatic backup
aipaca restore

# Preview restoration
aipaca restore --dry-run

# Restore from a specific backup
aipaca restore --backup myrepo-2024-01-15-143022
```

### `aipaca clean [repo-path]`

Remove AI files from repository (with backup).

```bash
# Clean current repo
aipaca clean

# Preview what would be removed
aipaca clean --dry-run

# Clean without backup (dangerous!)
aipaca clean --no-backup
```

**Use case**: Creating clean commits or PRs without AI configuration files.

```bash
# Workflow for clean commits
aipaca clean
git add . && git commit -m "feat: new feature"
git push
aipaca restore  # Bring back AI files 🦙
```

### `aipaca status [repo-path]`

Show current state of AI files.

```bash
aipaca status
```

Output:
```
Repo: /Users/you/project

Applied profile: default (modified)
Applied at: 2024-01-15 14:30:22
Backup: project-2024-01-15-143022

AI files in repo:
  .claude/ (5 files)
  .cursor/ (3 files)
  CLAUDE.md

Changes since apply:
  M .claude/agents/custom.md
  A .claude/agents/new-agent.md

Available backups (2):
  project-2024-01-15-143022 (9 files)
  project-2024-01-14-091533 (8 files)
```

### `aipaca diff [profile] [repo-path]`

Show differences between repo and profile.

```bash
# Compare against currently applied profile
aipaca diff

# Compare against specific profile
aipaca diff minimal
```

Output:
```
Comparing against profile 'default'

Changes:
  + .claude/agents/new-agent.md (added in repo)
  - .claude/agents/old-agent.md (missing from repo)
  M .claude/agents/custom.md (modified)
```

### `aipaca profiles`

Manage profiles.

```bash
# List all profiles
aipaca profiles list

# Show profile contents
aipaca profiles show default

# Copy a profile
aipaca profiles copy default my-backup

# Delete a profile
aipaca profiles delete old-profile
```

**List output:**
```
PROFILE       DESCRIPTION                    FILES
-------       -----------                    -----
default       Standard AI setup              12
minimal       Lightweight config             4
experimental  Testing new prompts            8
```

## Configuration

Configuration is stored at `~/.aipaca.yaml`:

```yaml
version: "1"

storage:
  path: "~/.aipaca"

# Patterns that define "AI files"
ai_patterns:
  - ".claude"
  - ".claude/**"
  - ".cursor"
  - ".cursor/**"
  - "CLAUDE.md"
  - "ai/"
  - "ai/**"
  - ".ai*"

# Default profile when none specified
default_profile: "default"

# Optional profile descriptions
profile_descriptions:
  default: "Standard AI setup with Claude and Cursor"
  minimal: "Lightweight Claude-only configuration"
```

## Storage Structure

```
~/.aipaca/
├── profiles/                    # Your AI config library (the herd 🦙)
│   ├── default/
│   │   ├── .claude/
│   │   │   ├── agents/
│   │   │   └── commands/
│   │   ├── .cursor/
│   │   └── CLAUDE.md
│   ├── minimal/
│   └── experimental/
│
├── backups/                     # Automatic backups
│   ├── myrepo-2024-01-15-143022/
│   └── myrepo-2024-01-14-091533/
│
└── state/
    └── repo-states.yaml         # Tracks what's applied where
```

## Workflows

### Workflow 1: Consistent Setup Across Projects

```bash
# Set up your "golden" configuration once
cd ~/my-main-project
aipaca init --import  # Creates "default" profile

# Apply to other projects
cd ~/other-project
aipaca apply default

cd ~/another-project
aipaca apply default
```

### Workflow 2: Experimenting with New Prompts

```bash
# Save current working setup
aipaca save --as stable

# Apply experimental profile
aipaca apply experimental

# Test and iterate...
# If it works, save it
aipaca save

# If it doesn't work, restore
aipaca restore
```

### Workflow 3: Project-Specific Configurations

```bash
# Start with base config
aipaca apply default

# Customize for this project...
# Then save as project-specific profile
aipaca save --as project-webapp

# Later, in another clone of same project
aipaca apply project-webapp
```

### Workflow 4: Clean Commits

```bash
# Remove AI files before committing
aipaca clean

# Make your commit
git add .
git commit -m "feat: implement feature"
git push

# Restore AI files
aipaca restore
```

### Workflow 5: Sharing Configurations

```bash
# Your profiles are just directories
ls ~/.aipaca/profiles/

# Share a profile
cp -r ~/.aipaca/profiles/my-setup ~/shared/

# Import a shared profile
cp -r ~/shared/team-setup ~/.aipaca/profiles/
```

### Workflow 6: First Time Setup (Personal Config for Team Repo)

When you want to use your own AI configs while working on a team repo that has its own AI files:

```bash
# Pull latest from remote
git pull origin main

# Save repo's AI files as "original" (so you can restore later)
aipaca save --as original

# Now create/modify your own AI configs...
# (edit .claude/, CLAUDE.md, etc)

# Save YOUR configs as a profile
aipaca save --as my-config

# Restore repo's original AI files
aipaca apply original

# Commit and push (with original AI files, not yours)
git add . && git commit -m "your changes" && git push
```

### Workflow 7: Daily Work with Personal Config

After your profile exists, use this workflow daily:

```bash
# Pull latest
git pull origin main

# Save whatever AI files came from remote (in case they changed)
aipaca save --as original --force

# Apply your personal config
aipaca apply my-config

# ✨ Work with your AI setup ✨

# See what you changed vs your profile
aipaca diff

# Option A: Update existing profile with changes
aipaca save

# Option B: Save as new profile
aipaca save --as my-config-v2

# Restore original repo AI files before committing
aipaca apply original

# Commit and push
git add . && git commit -m "your changes" && git push
```

### Quick Reference

| What you want | Command |
|---------------|---------|
| Save repo's AI files | `aipaca save --as original` |
| Apply your config | `aipaca apply my-config` |
| See what changed | `aipaca diff` |
| Update your profile | `aipaca save` |
| Save as new profile | `aipaca save --as new-name` |
| Restore original | `aipaca apply original` |
| Preview any action | add `--dry-run` |

## Safety Features

1. 🔒 **Automatic Backups**: Every `apply` and `clean` creates a timestamped backup
2. 👀 **Dry Run Mode**: Preview any operation with `--dry-run`
3. 📍 **State Tracking**: Know exactly what profile is applied where
4. ✅ **Checksums**: File integrity verification during operations
5. ⚠️ **Confirmation Prompts**: Destructive operations require confirmation

## Supported AI Tools

aipaca works with any file-based AI configuration, including:

- **Claude Code** (`.claude/`, `CLAUDE.md`)
- **Cursor** (`.cursor/`)
- **GitHub Copilot** (`.github/copilot/`)
- **Continue.dev** (`.continue/`)
- **Aider** (`.aider/`)
- **Custom AI tools** (configure patterns in `~/.aipaca.yaml`)

## Why "aipaca"?

Because managing AI configs should be as pleasant as hanging out with alpacas! 🦙

Also: **AI** + al**paca** = **aipaca**

## Requirements

- Go 1.21 or later (for building from source)
- macOS, Linux, or Windows

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [doublestar](https://github.com/bmatcuk/doublestar) for glob patterns
- Inspired by the need to stop copy-pasting AI configs everywhere

---

Made with 🦙 by developers who got tired of managing AI configs manually.
