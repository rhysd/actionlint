package actionlint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func ExampleCommand() {
	var output bytes.Buffer

	// Create command instance populating stdin/stdout/stderr
	cmd := Command{
		Stdin:  os.Stdin,
		Stdout: &output,
		Stderr: &output,
	}

	// Run the command end-to-end. Note that given args should contain program name
	workflow := filepath.Join(".github", "workflows", "release.yaml")
	status := cmd.Main([]string{"actionlint", "-shellcheck=", "-pyflakes=", workflow})

	fmt.Println("Exited with status", status)
	// Output: Exited with status 0

	if status != 0 {
		panic("actionlint command failed: " + output.String())
	}
}
