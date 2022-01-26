package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gophers.dev/cmds/new-cl/tool"
	"gophers.dev/pkgs/extractors/env"
)

const usage = `
usage)
  %s [type] [pr] <message>

args)
  type:    %s
  pr:      # from github
  message: (optional) directly insert message in note

example)
  %s bug 11235 "Fixed a bug"
`

func outputUsage(w io.Writer, name string) {
	_, _ = io.WriteString(w, fmt.Sprintf(
		usage, name, strings.ReplaceAll(
			tool.ChangelogKinds, ",", "|",
		), name,
	))
}

func outputErr(w io.Writer, err error) {
	_, _ = io.WriteString(w, fmt.Sprintf("%s\n", err.Error()))
}

func main() {
	r := &tool.Runner{
		Output: os.Stdout,
		Env:    env.OS,
		Args:   os.Args[1:],
	}

	err := r.Run()
	switch {
	case errors.Is(tool.ArgErr, err):
		outputUsage(os.Stderr, os.Args[0])
		os.Exit(1)
	case err != nil:
		outputErr(os.Stderr, err)
		os.Exit(1)
	}
}
