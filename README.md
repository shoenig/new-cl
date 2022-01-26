# new-cl

Create Changelog entries consumed by Hashicorp's [go-changelog](https://github.com/hashicorp/go-changelog)

# Getting Started

The `new-cl` command can be installed by running

```bash
go install gophers.dev/cmds/new-cl@latest
```

# Optional Configuration

`CHANGELOG_DIR` environment variable

By default `new-cl` looks for a `.changelog` directory for storage of changelog notes.
A different directory name can be specified by setting `CHANGELOG_DIR` to a different name.

# Usage

```
new-cl [type] [pr] <message>
```

`type`: one of `bug`, `improvement`, `security`, `breaking-change`, `deprecation`, `note`
`pr`: the PR number assigned by GitHub
`message`: (optional) the message the changelog note will contain

# Example Usages

```bash
new-cl bug 11235 "Fixed a bug"
```

# Contributing

The `gophers.dev/cmds/new-cl` module is always improving with new features and
error corrections. For contributing bug fixes and new features please file an
issue.

# License

The `gophers.dev/cmds/new-cl` module is open source under the [BSD](LICENSE)
