package tool

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gophers.dev/pkgs/extractors/env"
	"gophers.dev/pkgs/ignore"
)

const (
	changelogDirEnv = "CHANGELOG_DIR"
	ChangelogDir    = ".changelog"

	changelogKindEnv = "CHANGELOG_KINDS"
	ChangelogKinds   = "bug,improvement,security,breaking-change,deprecation,note"
)

type Runner struct {
	Output io.Writer
	Env    env.Environment
	Args   []string
}

func (r *Runner) Run() error {
	params, err := r.extractArgs()
	if err != nil {
		return err
	}

	targetDir, err := r.getChangelogDir()
	if err != nil {
		return err
	}

	actualDir, err := findTargetDir(targetDir)
	if err != nil {
		return err
	}

	file, err := createFile(actualDir, params)
	if err != nil {
		return err
	}

	_, err = io.WriteString(
		r.Output,
		fmt.Sprintf("created note: %s\n", filepath.Base(file)),
	)
	return err
}

// findTargetDir allows the user to run from within the .changelog directory, or
// from the parent of that directory (for convenience). The path returned is the
// actual path of the .changelog directory, in which the note file will be created.
func findTargetDir(targetDir string) (string, error) {
	full, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	// are we in the CL dir?
	cwd := filepath.Base(full)
	if cwd == targetDir {
		return ".", nil
	}

	// is the CL dir a subdirectory?
	sub := filepath.Join(full, targetDir)
	_, statErr := os.Stat(sub)
	return sub, statErr
}

func (r *Runner) getChangelogDir() (string, error) {
	changelogDir := ChangelogDir
	if err := env.Parse(r.Env, env.Schema{
		changelogDirEnv: env.String(&changelogDir, false),
	}); err != nil {
		return "", err
	}
	return changelogDir, nil
}

func (r *Runner) extractArgs() (*Params, error) {
	if err := checkNumArgs(len(r.Args)); err != nil {
		return nil, err
	}

	kind := r.Args[0]
	if err := r.checkKind(kind); err != nil {
		return nil, err
	}

	pr, err := strconv.Atoi(r.Args[1])
	if err != nil {
		return nil, fmt.Errorf("pr must be a number")
	}

	var note string
	if len(r.Args) == 3 {
		note = r.Args[2]
	}

	return &Params{
		Type: kind,
		PR:   pr,
		Note: note,
	}, nil
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

	s := fmt.Sprintf("```release-note:%s\n%s\n```\n", p.Type, note)
	_, err := io.WriteString(w, s)
	return err
}

var ArgErr = errors.New("number of arguments")

func checkNumArgs(n int) error {
	switch n {
	case 2, 3:
		return nil
	}
	return ArgErr
}

func (r *Runner) checkKind(s string) error {
	l := strings.ToLower(s)

	kinds := ChangelogKinds
	if err := env.Parse(r.Env, env.Schema{
		changelogKindEnv: env.String(&kinds, false),
	}); err != nil {
		return err
	}

	for _, kind := range strings.Split(kinds, ",") {
		if l == kind {
			return nil
		}
	}
	return fmt.Errorf("unknown kind %q", s)
}

func createFile(dir string, p *Params) (string, error) {
	path := filepath.Join(dir, p.Filename())
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer ignore.Close(f)
	if wErr := p.Write(f); wErr != nil {
		return "", wErr
	}
	return path, nil
}
