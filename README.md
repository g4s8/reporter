Generates reports for GitHub repositories

## Install

Download binary for your platform from GitHub releases:
https://github.com/g4s8/reporter/releases

## Usage

Create API token without permissions (to increase GitHub API quota limits).
To use this token with reporter:
 1. Put it to `~/.config/reporter/github_token.txt`
 2. Set it to `GITHUB_TOKEN` environment variable
 3. Use `--token` CLI option

The syntax is: `report <action> <source>` (see `report --help` for details)
where actions is either `pulls` or (will be later),
and source is either organization name (for full report over all repositories)
or full repository coordinates (`user/repo`).
This is a few valid examples:
 - `report pulls artipie`
 - `report pulls cqfn/diktat`
 - `report pulls cqfn/degitx cqfn/degitx-simulator`

See `report pulls --help` for more details
