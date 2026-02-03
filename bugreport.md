# Bug Report: `go-subcommand` generation issue

## Description
When generating the CLI code, the `root.go` file is missing the import for the `commands` package and the `commands.` prefix when calling the root command function.

## Steps to Reproduce
1.  Define a root command in `commands/root.go` as `func PodcastCdrManager(profile string)`.
2.  Run `gosubc generate --dir ..`.
3.  Inspect `cmd/podcast-cdr-manager/root.go`.

## Expected Behavior
`root.go` should import `github.com/arran4/podcast-cdr-manager/commands` and call `commands.PodcastCdrManager(c.profile)`.

## Actual Behavior
`root.go` does not import the package and calls `PodcastCdrManager(c.profile)` directly, leading to compilation errors.

## Workaround
Manually patch `root.go` to add the import and correct the function call.
