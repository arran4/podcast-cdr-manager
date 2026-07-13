# podcast-cdr-manager Skill for AI Agents

This document provides operational guidance for AI agents interacting with the `podcast-cdr-manager` CLI application.

## Overview
`podcast-cdr-manager` is a CLI tool designed to manage podcast subscriptions and create ISO images (for burning to CDROMs, CDR, CDRW). It tracks played episodes, fetches new ones, and packs unplayed ones into 600MB ISO disks.

## Core Concepts
- **Profiles**: The app stores data by "profile" using the XDG Base Directory specification (e.g., `~/.config/podcast-cdr-manager/default.yaml`). The default profile is `default`. You often need to specify a profile using `--profile` if not using the default, but commands default to `default` automatically. If the app is run in dev mode, it might use `default-dev`.
- **Subscriptions**: RSS feed URLs containing podcast episodes.
- **Casts**: Individual podcast episodes (MP3 files).
- **Disks**: Virtual representations of CDs. A disk is filled with casts up to roughly 600MB, after which an ISO can be generated.

## Command Reference and Agent Guidance

### Profile Creation
- **Command**: `podcast-cdr-manager profile new <name>`
- **Note**: A profile must be created before most other operations can succeed.
- **Agent Tip**: Always ensure a profile exists before trying to subscribe or list casts. If testing, run `podcast-cdr-manager profile new test` first.

### Managing Subscriptions
- **Command**: `podcast-cdr-manager subscribe rss <url>`
- **Command**: `podcast-cdr-manager subscriptions list-subs`
- **Command**: `podcast-cdr-manager subscriptions refresh`
- **Agent Tip**: `subscriptions refresh` goes through all feeds, checks for updates, and adds the results to the unused podcast list. This is a crucial step after subscribing or after some time has passed.

### Managing Disks and ISOs
- **Command**: `podcast-cdr-manager disk next -create -dry=false`
- **Dry-run Gotcha**: The `-dry` flag defaults to `true` on disk operations. **You must explicitly pass `-dry=false` if you want changes to be saved to the profile.**
- **Command**: `podcast-cdr-manager disk iso generate -dry=false -index 0`
- **Agent Tip**: Like `disk next`, `disk iso generate` has `-dry=true` by default. It downloads MP3s and generates the ISO. Ensure `-dry=false` is used when actually producing the ISO.

### Listing Data
- **Command**: `podcast-cdr-manager cast list-casts`
- **Command**: `podcast-cdr-manager disk list-disks`
- **Agent Tip**: Output for these commands is formatted text (tables). When parsing output programmatically, note the column alignments.

### Skill Management
This CLI includes a built-in skill management system (`podcast-cdr-manager skill`) designed to install tools like this very document.
- **Install**: `podcast-cdr-manager skill install <source>` (e.g., `<app> skill install owner/repo`)
- **Update**: `podcast-cdr-manager skill update <name>`
- **Remove**: `podcast-cdr-manager skill remove <name>`
- **List**: `podcast-cdr-manager skill list` (supports `--json`)
- **Agent Tip**: Skills are tracked with metadata to ensure reproducible updates without overwriting local modifications unless `--force` is provided.

## Common Pitfalls for AI Agents
1. **Forgetting `-dry=false`**: As noted above, the disk allocation and ISO generation commands are safe by default. If you are trying to mutate state and it's not sticking, you forgot `-dry=false`.
2. **Missing Profiles**: Attempting to list casts or disks before a profile exists will result in an error reading the config file.
3. **Empty RSS Feeds**: If `subscribe rss` fails or adds 0 items, check if the RSS URL is reachable and correctly formatted.

## Environment Variables
The application uses the `xdg` library, so it respects standard XDG environment variables like `XDG_CONFIG_HOME`.
