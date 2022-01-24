package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gophers.dev/pkgs/ignore"
)

const changelogDir = ".changelog"

// pulled from nomad .changelog/changelog.tmpl
var kinds = []string{
	"breaking-change",
	"security",
	"improvement",
	"deprecation",
	"bug",
	"note",
}

func main() {
	dir, err := findCLDir(changelogDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "must run in %s: %s", changelogDir, err)
		os.Exit(1)
	}

	params, err := extract(os.Args[1:])
	if err != nil {
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
	fmt.Println("cwd:", cwd, "required:", required)
	if cwd == required {
		return ".", nil
	}

	// is the CL dir a subdirectory?
	sub := filepath.Join(full, changelogDir)
	fmt.Println("sub:", sub)
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
}

func (p *Params) Filename() string {
	return fmt.Sprintf("%d.txt", p.PR)
}

func (p *Params) Write(w io.Writer) error {
	s := fmt.Sprintf("```\nrelease-note:%s\nMESSAGE\n```\n", p.Type)
	_, err := io.WriteString(w, s)
	return err
}

func extract(args []string) (*Params, error) {
	if n := len(args); n != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", n)
	}

	kind := args[0]
	if err := checkKind(kind); err != nil {
		return nil, err
	}

	pr, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("pr must be a number")
	}

	return &Params{
		Type: kind,
		PR:   pr,
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
