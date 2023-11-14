package actionlint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFilesAsWorkflow(t *testing.T) {
	dir := filepath.Join("testdata", "detect_format", "workflows")
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		_, _, sf, _ := ParseFile(path, b, FileAutoDetect)
		if sf != FileWorkflow {
			t.Errorf("file %s is not detected as workflow", path)
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("could not read files in %q: %w", dir, err))
	}
}

func TestDetectFilesAsAction(t *testing.T) {
	dir := filepath.Join("testdata", "detect_format", "actions")
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		_, _, sf, _ := ParseFile(path, b, FileAutoDetect)
		if sf != FileAction {
			t.Errorf("file %s is not detected as action", path)
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("could not read files in %q: %w", dir, err))
	}
}
