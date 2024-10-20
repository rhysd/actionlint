package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/rhysd/actionlint"
)

func GeneratePermalink(src []byte) (string, error) {
	var out bytes.Buffer

	b64 := base64.NewEncoder(base64.StdEncoding, &out)
	comp, _ := zlib.NewWriterLevel(b64, zlib.BestCompression)

	scan := bufio.NewScanner(bytes.NewReader(src))
	first := true
	for scan.Scan() {
		l := scan.Bytes()
		if bytes.HasPrefix(bytes.TrimSpace(l), []byte("#")) {
			continue
		}
		if first {
			first = false
		} else {
			comp.Write([]byte{'\n'})
		}
		comp.Write(l)
	}
	if err := scan.Err(); err != nil {
		return "", err
	}

	comp.Close()
	b64.Close()

	return fmt.Sprintf("[Playground](https://rhysd.github.io/actionlint/#%s)", out.Bytes()), nil
}

func Actionlint(src []byte) ([]byte, error) {
	var out bytes.Buffer

	opts := &actionlint.LinterOptions{
		StdinFileName: "test.yaml",
		Shellcheck:    "shellcheck",
		Pyflakes:      "pyflakes",
		Color:         actionlint.ColorOptionKindNever,
	}

	l, err := actionlint.NewLinter(&out, opts)
	if err != nil {
		return nil, err
	}

	p, err := actionlint.NewProjects().At(".")
	if err != nil {
		return nil, err
	}
	if _, err := l.Lint("test.yaml", src, p); err != nil {
		return nil, err
	}

	// Some error message contains absolute file paths. Replace them not to make the document depend
	// on the current file system.
	b := bytes.ReplaceAll(out.Bytes(), []byte(p.RootDir()), []byte("/path/to/repo"))

	return b, nil
}

func Update(in []byte) ([]byte, error) {
	var buf bytes.Buffer

	var input bytes.Buffer
	var anchor string
	var section string
	var inputHeader bool
	var outputHeader bool
	var inInput bool
	var skipOutput bool
	var count int
	lnum := 0
	scan := bufio.NewScanner(bytes.NewReader(in))
	for scan.Scan() {
		lnum++
		l := scan.Text()
		if strings.HasPrefix(l, "## ") && anchor != "" {
			if input.Len() > 0 {
				return nil, fmt.Errorf("example input for %q was not consumed. playground link may not exist", section)
			}
			if section != "" && count == 0 {
				log.Printf("Section %q contains NO example", section)
			}
			section = l[3:]
			log.Printf("Entering new section %q (%s) at line %d", section, anchor, lnum)
			anchor = ""
			inputHeader = false
			outputHeader = false
			inInput = false
			skipOutput = false
			count = 0
		}
		if strings.HasPrefix(l, `<a name="`) && strings.HasSuffix(l, `"></a>`) {
			anchor = strings.TrimSuffix(strings.TrimPrefix(l, `<a name="`), `"></a>`)
		}
		if l == "Example input:" {
			log.Printf("Found example input header for %q at line %d", section, lnum)
			inputHeader = true
		}
		if l == "```yaml" && inputHeader {
			inputHeader = false
			inInput = true
		} else if inInput && l != "```" {
			input.WriteString(l)
			input.WriteByte('\n')
		}
		if l == "```" {
			if inInput {
				log.Printf("Found an input example (%d bytes) for %q at line %d", input.Len(), section, lnum)
				inInput = false
			} else if outputHeader {
				buf.WriteString("```\n")
				for {
					if !scan.Scan() {
						return nil, fmt.Errorf("code block for output of %q does not close", section)
					}
					lnum++
					if scan.Text() == "```" {
						break
					}
				}
				if input.Len() == 0 {
					return nil, fmt.Errorf("output cannot be generated because example input for %q does not exist", section)
				}
				log.Printf("Generating output for the input example for %q at line %d", section, lnum)
				out, err := Actionlint(input.Bytes())
				if err != nil {
					return nil, err
				}
				buf.Write(out)
			}
		}
		if l == "Output:" {
			log.Printf("Found example output header for %q at line %d", section, lnum)
			outputHeader = true
		}
		if l == "<!-- Skip update output -->" {
			log.Printf("Skip updating output for %q due to the comment at line %d", section, lnum)
			outputHeader = false
			skipOutput = true
		}
		if strings.HasPrefix(l, "[Playground](https://rhysd.github.io/actionlint/#") && strings.HasSuffix(l, ")") {
			if input.Len() == 0 {
				return nil, fmt.Errorf("playground link cannot be generated because example input for %q does not exist", section)
			}
			if !outputHeader && !skipOutput {
				return nil, fmt.Errorf("output code block is missing for %q", section)
			}
			link, err := GeneratePermalink(input.Bytes())
			if err != nil {
				return nil, err
			}
			buf.WriteString(link)
			buf.WriteByte('\n')
			log.Printf("Generate playground link for %q at line %d: %s", section, lnum, link)
			outputHeader = false
			input.Reset()
			count++
			continue
		}
		if l == "<!-- Skip playground link -->" {
			log.Printf("Skip generating playground link for %q due to the comment at line %d", section, lnum)
			outputHeader = false
			if input.Len() == 0 {
				return nil, fmt.Errorf("example input for %q is empty", section)
			}
			input.Reset()
			count++
		}
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	if inInput {
		return nil, fmt.Errorf("code block for example input for %q is not closed", section)
	}
	return buf.Bytes(), scan.Err()
}

func Main(args []string) error {
	var path string
	var check bool
	switch len(args) {
	case 2:
		path = args[1]
	case 3:
		if args[1] != "-check" && args[1] != "--check" {
			return errors.New("usage: update-checks-doc [-check] FILE")
		}
		path = args[2]
		check = true
	default:
		return errors.New("usage: update-checks-doc [-check] FILE")
	}

	in, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read the document file: %w", err)
	}
	log.Printf("Read %d bytes from %q", len(in), path)

	out, err := Update(in)
	if err != nil {
		return err
	}

	if bytes.Equal(in, out) {
		log.Printf("Do nothing because there is no update in %q", path)
		return nil
	}

	if check {
		return errors.New("checks document has some update. run `go run ./scripts/update-checks-doc ./docs/checks.md` and commit the changes. the diff:\n\n" + cmp.Diff(in, out))
	}

	log.Printf("Generate the updated content (%d bytes) for %q", len(out), path)
	return os.WriteFile(path, out, 0666)
}

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
