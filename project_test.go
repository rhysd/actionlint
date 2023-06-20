package actionlint

import (
	"os"
	"path/filepath"
	"testing"
)

// Create `.git` directory since actionlint finds the directory to detect the repository root.
// Without creating this directory, this test case will fail when `actionlint/.git` directory
// doesn't exist. When cloning actionlint repository with Git, it never happens. However, when
// downloading sources tarball from github.com, it doesn't contain `.git` directory so it
// happens. Please see #307 for more details.
func testEnsureDotGitDir(dir string) {
	d := filepath.Join(dir, ".git")
	if err := os.MkdirAll(d, 0750); err != nil {
		panic(err)
	}
}

func TestProjectsFindProjectFromPath(t *testing.T) {
	d := filepath.Join("testdata", "find_project")
	abs, err := filepath.Abs(d)
	if err != nil {
		panic(err)
	}
	testEnsureDotGitDir(d)

	ps := NewProjects()
	for _, tc := range []struct {
		what string
		path string
	}{
		{
			what: "project root",
			path: d,
		},
		{
			what: "workflows directory",
			path: filepath.Join(d, ".github", "workflows"),
		},
		{
			what: "workflow file",
			path: filepath.Join(d, ".github", "workflows", "test.yaml"),
		},
		{
			what: "outside workflows directory",
			path: filepath.Join(d, ".github", "reusable", "broken.yaml"),
		},
		{
			what: "directory outside .github",
			path: filepath.Join(d, "foo"),
		},
		{
			what: "file outside .github",
			path: filepath.Join(d, "foo", "test.txt"),
		},
	} {
		t.Run(tc.what, func(t *testing.T) {
			p := ps.At(tc.path)

			r := p.RootDir()
			if r != abs {
				t.Fatalf("root directory of project %v should be %q but got %q", p, abs, r)
			}

			// Result should be cached
			p2 := ps.At(tc.path)
			if p != p2 {
				t.Fatalf("project %v is not cached. New project is %v. %p v.s. %p", p, p2, p, p2)
			}
		})
	}
}

func TestProjectsDoesNotFindProjectFromOutside(t *testing.T) {
	d := filepath.Join("testdata", "find_project")
	abs, err := filepath.Abs(d)
	if err != nil {
		panic(err)
	}
	testEnsureDotGitDir(d)

	outside := filepath.Join(d, "..")
	ps := NewProjects()
	p := ps.At(outside)
	if p != nil && p.RootDir() == abs {
		t.Fatalf("project %v is detected from outside of the project %q", p, outside)
	}
}
