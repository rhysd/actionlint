package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/rhysd/actionlint"
)

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
	errs, err := l.Lint("test.yaml", src, p)
	if err != nil {
		return nil, err
	}
	if len(errs) == 0 {
		return nil, errors.New("the input example caused no error")
	}

	// Some error message contains absolute file paths. Replace them not to make the document depend
	// on the current file system.
	b := bytes.ReplaceAll(out.Bytes(), []byte(p.RootDir()), []byte("/path/to/repo"))

	return b, nil
}

type state int

const (
	stateInit state = iota
	stateAnchor
	stateHeading
	stateInputHeader
	stateInputBlock
	stateAfterInput
	stateOutputHeader
	stateAfterOutput
	stateOutputBlock
	stateEnd
)

func (s state) String() string {
	switch s {
	case stateAnchor:
		return "anchor"
	case stateHeading:
		return "heading"
	case stateInputHeader:
		return "input header"
	case stateInputBlock:
		return "input block"
	case stateAfterInput:
		return "after input"
	case stateOutputHeader:
		return "output header"
	case stateAfterOutput:
		return "after output"
	case stateOutputBlock:
		return "output block"
	case stateEnd:
		return "end"
	default:
		return "init"
	}
}

type Updater struct {
	prev, cur state
	lines     *bufio.Scanner
	out       bytes.Buffer
	heading   string
	input     bytes.Buffer
	lnum      int
	ids       map[string]int
	headings  map[string]int
	firstErr  error
}

func NewUpdater(in []byte) *Updater {
	return &Updater{
		lines:    bufio.NewScanner(bytes.NewReader(in)),
		ids:      map[string]int{},
		headings: map[string]int{},
	}
}

func (u *Updater) err(err error) {
	if u.firstErr == nil && err != nil {
		u.firstErr = fmt.Errorf("error at line %d while generating section %q: %w", u.lnum, u.heading, err)
	}
}

func (u *Updater) End() ([]byte, error) {
	u.err(u.lines.Err())
	if u.cur != stateEnd {
		u.err(fmt.Errorf("unexpected state %q after generating all. this happens when a block is unclosed or some part was missing", u.cur))
	}
	log.Printf("Scanned %d lines and %d sections", u.lnum, len(u.headings))
	return u.out.Bytes(), u.firstErr
}

func (u *Updater) PlaygroundLink(src []byte) string {
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
	u.err(scan.Err())

	u.err(comp.Close())
	u.err(b64.Close())

	return fmt.Sprintf("[Playground](https://rhysd.github.io/actionlint/#%s)", out.Bytes())
}

func (u *Updater) state(s state, reason string) {
	log.Printf("%s at line %d in section %q (%q -> %q)", reason, u.lnum, u.heading, u.cur, s)
	u.prev, u.cur = u.cur, s
}

func (u *Updater) Scan() bool {
	if u.firstErr != nil || !u.lines.Scan() {
		return false
	}
	u.lnum++
	return true
}

func (u *Updater) expect(states ...state) {
	for _, s := range states {
		if s == u.cur {
			return
		}
	}
	u.err(fmt.Errorf("unexpected state %q. expected %q", u.cur, states))
}

func (u *Updater) Update() {
	l := u.lines.Text()

	isHeading := strings.HasPrefix(l, "## ")
	isInputHeader := l == "Example input:"
	isOutputHeader := l == "Output:"
	isSkipOutput := l == "<!-- Skip update output -->"
	isSkipPlaygroundLink := l == "<!-- Skip playground link -->"
	isPlaygroundLink := strings.HasPrefix(l, "[Playground](") && strings.HasSuffix(l, ")")

	// Validation
	switch {
	case isHeading:
		u.expect(stateAnchor)
	case isInputHeader:
		u.expect(stateHeading, stateEnd)
	case isOutputHeader:
		u.expect(stateAfterInput)
	case isSkipOutput:
		u.expect(stateOutputHeader)
	case isSkipPlaygroundLink, isPlaygroundLink:
		u.expect(stateAfterOutput)
	}

	// State transition
	//
	//    backtrack
	//    ┌───────┐
	// ┌──▼─┐  ┌──┴───┐  ┌───────┐  ┌──────┐  ┌─────┐  ┌─────┐
	// │init│─►│anchor│─►│heading│─►│input │─►│input│─►│after│────┐
	// └────┘  │      │  │       │  │header│  │block│  │input│    │
	//         └──▲┬──┘  └───────┘  └──▲───┘  └─────┘  └─────┘ ┌──▼───┐
	//     next   ││            more   │            skip       │output│
	//     section││            example│        ┌──────────────│header│
	//            ││backtrack        ┌───┐  ┌───▼──┐  ┌──────┐ └──┬───┘
	//            │└────────────────►│end│◄─│after │◄─│output│◄───┘
	// ┌────┐     └──────────────────│   │  │output│  │block │
	// │done│◄───────────────────────│   │  └──────┘  └──────┘
	// └────┘                        └───┘
	//
	switch u.cur {
	case stateInit, stateEnd:
		if u.cur == stateEnd && isInputHeader {
			u.state(stateInputHeader, "Found more example input")
			break
		}
		if strings.HasPrefix(l, `<a id="`) && strings.HasSuffix(l, `"></a>`) {
			id := strings.TrimSuffix(strings.TrimPrefix(l, `<a id="`), `"></a>`)
			if len(id) == 0 {
				u.err(errors.New("id for <a> tag is empty"))
				return
			}
			if n, ok := u.ids[id]; ok {
				u.err(fmt.Errorf("id %q was already used at line %d", id, n))
				return
			}
			u.ids[id] = u.lnum
			u.state(stateAnchor, "Found new <a> ID "+id)
		}
	case stateAnchor:
		if isHeading {
			h := strings.TrimPrefix(l, "## ")
			if n, ok := u.headings[h]; ok {
				u.err(fmt.Errorf("heading %q was already used at line %d", h, n))
				return
			}
			u.headings[h] = u.lnum
			u.heading = h
			u.state(stateHeading, "Entering new section")
		} else {
			u.state(u.prev, "Back to previous state because the next line to <a> is not a section heading")
			u.Update()
			return
		}
	case stateHeading:
		if isInputHeader {
			u.state(stateInputHeader, "Found example input header")
		}
	case stateInputHeader:
		if l == "```yaml" {
			u.state(stateInputBlock, "Start code block for input example")
		}
	case stateInputBlock:
		if l == "```" {
			if u.input.Len() == 0 {
				u.err(errors.New("empty example input is not allowed"))
				return
			}
			u.state(stateAfterInput, "End code block for input example")
		} else {
			u.input.WriteString(l)
			u.input.WriteByte('\n')
		}
	case stateAfterInput:
		if isOutputHeader {
			u.state(stateOutputHeader, "Found example output header")
		}
	case stateOutputHeader:
		if isSkipOutput {
			u.state(stateAfterOutput, "Skip updating output due to the comment")
		} else if l == "```" {
			u.state(stateOutputBlock, "Start code block for output")
		}
	case stateOutputBlock:
		if l != "```" {
			return // Output block will be generated by actionlint
		}
		out, err := Actionlint(u.input.Bytes())
		u.err(err)
		u.out.Write(out)
		u.state(stateAfterOutput, "End code block for output and generate actionlint output")
	case stateAfterOutput:
		if isSkipPlaygroundLink {
			u.input.Reset()
			u.state(stateEnd, "Skip updating playground link due to the comment")
		} else if isPlaygroundLink {
			ln := u.PlaygroundLink(u.input.Bytes())
			u.out.WriteString(ln)
			u.out.WriteByte('\n')
			u.input.Reset()
			u.state(stateEnd, "Generate playground link "+ln)
			return
		}
	}

	u.out.WriteString(l)
	u.out.WriteByte('\n')
}

func Update(in []byte) ([]byte, error) {
	u := NewUpdater(in)
	for u.Scan() {
		u.Update()
	}
	return u.End()
}

var stderr io.Writer = os.Stderr

func Main(args []string) error {
	var check bool
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.BoolVar(&check, "check", false, "check the document is up-to-date")
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: update-checks-doc [FLAGS] FILE\n\nFlags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("this command should take exact one file path but got %v", flags.Args())
	}
	path := flags.Arg(0)

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
		return fmt.Errorf("checks document has some update. run `go run ./scripts/update-checks-doc %s` and commit the changes. the diff:\n\n%s", path, cmp.Diff(in, out))
	}

	log.Printf("Overwrite the file with the updated content (%d bytes) at %q", len(out), path)
	return os.WriteFile(path, out, 0666)
}

func main() {
	if err := Main(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
