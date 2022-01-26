package tool

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gophers.dev/pkgs/extractors/env"
)

func setup(t *testing.T, clDir string) string {
	tmpDir, err := ioutil.TempDir("", "new-cl-values")
	require.NoError(t, err)

	clPath := filepath.Join(tmpDir, clDir)
	err = os.Mkdir(clPath, 0770)
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	return tmpDir
}

func cleanup(t *testing.T, path string) {
	err := os.RemoveAll(path)
	require.NoError(t, err)
}

func Test_findTargetDir(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(curDir)
	}()

	root := setup(t, ".changelog")
	defer cleanup(t, root)

	t.Run("missing", func(t *testing.T) {
		_, err = findTargetDir("_changelog")
		require.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("exists", func(t *testing.T) {
		path, findErr := findTargetDir(".changelog")
		require.NoError(t, findErr)
		require.Equal(t, ".changelog", filepath.Base(path))
		require.Equal(t, byte('/'), path[0]) // abs path when using parent dir
	})

	t.Run("internal", func(t *testing.T) {
		chErr := os.Chdir(filepath.Join(root, ".changelog"))
		require.NoError(t, chErr)
		path, findErr := findTargetDir(".changelog")
		require.NoError(t, findErr)
		require.Equal(t, ".", filepath.Base(path)) // local relative when local
	})
}

func TestRunner_getChangelogDir(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		environ := env.NewEnvironmentMock(t)
		environ.GetenvMock.When(changelogDirEnv).Then("")

		dir, err := (&Runner{
			Env: environ,
		}).getChangelogDir()

		require.NoError(t, err)
		require.Equal(t, ".changelog", dir)

		environ.MinimockFinish()
	})

	t.Run("custom", func(t *testing.T) {
		environ := env.NewEnvironmentMock(t)
		environ.GetenvMock.When(changelogDirEnv).Then(".custom")

		dir, err := (&Runner{
			Env: environ,
		}).getChangelogDir()

		require.NoError(t, err)
		require.Equal(t, ".custom", dir)

		environ.MinimockFinish()
	})
}

func TestRunner_extractArgs(t *testing.T) {
	environ := env.NewEnvironmentMock(t)
	environ.GetenvMock.When(changelogKindEnv).Then(ChangelogKinds)

	try := func(args []string, expParams *Params, expErr error) {
		params, err := (&Runner{
			Args: args,
			Env:  environ,
		}).extractArgs()

		if expErr != nil {
			require.Nil(t, params)
			require.Equal(t, expErr, err)
		} else {
			require.Equal(t, expParams, params)
		}
	}

	// too few arguments
	try(nil, nil, ArgErr)

	// too many arguments
	try([]string{"1", "2", "3", "4"}, nil, ArgErr)

	// not acceptable kind
	try([]string{"feature", "12345"}, nil, errors.New(`unknown kind "feature"`))

	// not a acceptable number
	try([]string{"bug", "xxx"}, nil, errors.New("pr must be a number"))

	// no message set
	try([]string{"bug", "12345"}, &Params{
		Type: "bug",
		PR:   12345,
		Note: "",
	}, nil)

	// with message set
	try([]string{"bug", "12345", "values: fixed a bug"}, &Params{
		Type: "bug",
		PR:   12345,
		Note: "values: fixed a bug",
	}, nil)

	environ.MinimockFinish()
}

func TestRunner_checkKind(t *testing.T) {
	tests := []struct {
		kinds  string
		values []string
		expOK  bool
	}{{
		kinds:  "",
		values: []string{"bug", "improvement", "security", "breaking-change", "deprecation", "note"},
		expOK:  true,
	}, {
		kinds:  "",
		values: []string{"foo", "bar", "baz"},
		expOK:  false,
	}, {
		kinds:  "feature,regression,test",
		values: []string{"feature", "regression", "test"},
		expOK:  true,
	}, {
		kinds:  "feature,regression,test",
		values: []string{"bug", "improvement", "security", "breaking-change", "deprecation", "note"},
		expOK:  false,
	}}

	for _, test := range tests {
		environ := env.NewEnvironmentMock(t)
		environ.GetenvMock.When(changelogKindEnv).Then(test.kinds)
		for _, value := range test.values {
			err := (&Runner{
				Env: environ,
			}).checkKind(value)
			if test.expOK {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, fmt.Sprintf("unknown kind %q", value))
			}
		}
		environ.MinimockFinish()
	}
}

func Test_createFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "new-cl")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	path, err := createFile(dir, &Params{
		Type: "bug",
		PR:   12345,
		Note: "test: Hello, world!",
	})
	require.NoError(t, err)
	require.Equal(t, "12345.txt", filepath.Base(path))

	b, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "```release-note:bug\ntest: Hello, world!\n```\n", string(b))
}

func TestRunner_Run(t *testing.T) {
	curDir, err := os.Getwd()
	require.NoError(t, err)
	defer require.NoError(t, os.Chdir(curDir))

	root := setup(t, ".changelog")
	defer cleanup(t, root)

	environ := env.NewEnvironmentMock(t)
	environ.GetenvMock.When(changelogDirEnv).Then(ChangelogDir)
	environ.GetenvMock.When(changelogKindEnv).Then(ChangelogKinds)

	out := bytes.NewBuffer([]byte{})

	err = (&Runner{
		Output: out,
		Env:    environ,
		Args:   []string{"improvement", "11358", `"test: Updated a dependency"`},
	}).Run()
	require.NoError(t, err)
	result := out.String()
	require.Equal(t, "created note: 11358.txt\n", result)

	environ.MinimockFinish()
}
