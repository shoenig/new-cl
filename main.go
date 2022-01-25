package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gophers.dev/pkgs/ignore"
)

const changelogDir = ".changelog"

// pulled from nomad/.changelog/changelog.tmpl
var kinds = []string{
	"bug",
	"improvement",
	"security",
	"breaking-change",
	"deprecation",
	"note",
}

const usage = `
usage: %s [type] [issue/pr] <message>

type:     %s
issue/pr: from github
message:  (optional) directly insert message in note
`

func outputUsage(w io.Writer, name string) {
	_, _ = io.WriteString(w, fmt.Sprintf(usage, name, strings.Join(kinds, "|")))
}

func main() {
	dir, err := findCLDir(changelogDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "must run in %s: %s", changelogDir, err)
		os.Exit(1)
	}

	program := os.Args[0]
	args := os.Args[1:]

	params, err := extract(args)
	switch {
	case errors.Is(err, argsErr):
		outputUsage(os.Stderr, program)
		os.Exit(1)
	case err != nil:
		_, _ = fmt.Fprintf(os.Stderr, "bad arguments: %s", err)
		os.Exit(1)
	}

	err = createFile(dir, params)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create note: %s", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stderr, "created: %s/%s", changelogDir, params.Filename())
}

func findCLDir(required string) (string, error) {
	full, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	// are we in the CL dir?
	cwd := filepath.Base(full)
	if cwd == required {
		return ".", nil
	}

	// is the CL dir a subdirectory?
	sub := filepath.Join(full, changelogDir)
	_, statErr := os.Stat(sub)
	return sub, statErr
}

func createFile(dir string, p *Params) error {
	path := filepath.Join(dir, p.Filename())
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer ignore.Close(f)
	if wErr := p.Write(f); wErr != nil {
		return wErr
	}
	return nil
}

type Params struct {
	Type string
	PR   int
	Note string
}

func (p *Params) Filename() string {
	return fmt.Sprintf("%d.txt", p.PR)
}

func (p *Params) Write(w io.Writer) error {
	note := "NOTE"
	if p.Note != "" {
		note = p.Note
	}

	s := fmt.Sprintf("```\nrelease-note:%s\n%s\n```\n", p.Type, note)
	_, err := io.WriteString(w, s)
	return err
}

var argsErr = errors.New("number of arguments")

func checkNumArgs(n int) error {
	switch n {
	case 2, 3:
		return nil
	}
	return argsErr
}

func extract(args []string) (*Params, error) {
	if err := checkNumArgs(len(args)); err != nil {
		return nil, err
	}

	kind := args[0]
	if err := checkKind(kind); err != nil {
		return nil, err
	}

	pr, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("pr must be a number")
	}

	var note string
	if len(args) == 3 {
		note = args[2]
	}

	return &Params{
		Type: kind,
		PR:   pr,
		Note: note,
	}, nil
}

func checkKind(s string) error {
	l := strings.ToLower(s)
	for _, kind := range kinds {
		if l == kind {
			return nil
		}
	}
	return fmt.Errorf("unknown kind %q", s)
}
