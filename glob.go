package actionlint

import (
	"fmt"
	"strings"
	"text/scanner"
)

// Note:
// - Broken pattern causes a syntax error
//   - '+' or '?' at top of pattern
//   - preceding character of '+' or '?' is special character like '+', '?', '*'
//   - Missing ] in [...] pattern like '[0-9'
//   - Missing end of range in [...] like '[0-]'
// - \ can escape special characters like '['. Otherwise \ is handled as normal character
// - invalid characters for Git ref names are not checked on GitHub Actions runtime
//   - `man git-check-ref-format` for more details
//   - \ is invalid character for ref names. it means that \ can be used only for escaping special chars

// InvalidGlobPattern is an error on invalid glob pattern.
type InvalidGlobPattern struct {
	// Message is a human readable error message.
	Message string
	// Column is a column number of the error in the glob pattern.
	Column int
}

func (err *InvalidGlobPattern) Error() string {
	return fmt.Sprintf("%d: %s", err.Column, err.Message)
}

type globValidator struct {
	isRef bool
	prec  bool
	errs  []InvalidGlobPattern
	scan  scanner.Scanner
}

func (v *globValidator) error(msg string) {
	v.errs = append(v.errs, InvalidGlobPattern{msg, v.scan.Pos().Column})
}

func (v *globValidator) unexpected(char rune, what, why string) {
	unexpected := "unexpected EOF"
	if char != scanner.EOF {
		unexpected = fmt.Sprintf("unexpected character %q", char)
	}

	while := ""
	if what != "" {
		while = fmt.Sprintf(" while checking %s", what)
	}

	v.error(fmt.Sprintf("invalid glob pattern. %s%s. %s", unexpected, while, why))
}

func (v *globValidator) invalidRefChar(char rune, why string) {
	msg := fmt.Sprintf("character %q is invalid for Git ref name. %s. see `man git-check-ref-format` for more details. note that regular expression is unavailable", char, why)
	v.error(msg)
}

func (v *globValidator) init(pat string) {
	v.errs = []InvalidGlobPattern{}
	v.prec = false
	v.scan.Init(strings.NewReader(pat))
	v.scan.Error = func(s *scanner.Scanner, m string) {
		v.error(fmt.Sprintf("error while scanning glob pattern %q: %s", pat, m))
	}
}

func (v *globValidator) validateNext() bool {
	c := v.scan.Next()

	switch c {
	case '\\':
		switch v.scan.Peek() {
		case '[', '?', '*':
			if v.isRef {
				v.invalidRefChar(v.scan.Peek(), "ref name cannot contain spaces, ~, ^, :, [, ?, *")
			}
			c = v.scan.Next() // eat escaped character
		case '+', '\\':
			c = v.scan.Next() // eat escaped character
		default:
			// file path can contain '\' (`mkdir 'foo\bar'` works)
			if v.isRef {
				v.invalidRefChar('\\', "only special characters [, ?, +, *, \\ can be escaped with \\")
				c = v.scan.Next()
			}
		}
		v.prec = true
	case '?':
		if !v.prec {
			v.unexpected('?', "? (zero or one of preceding character)", "the preceding character must not be special character")
		}
		v.prec = false
	case '+':
		if !v.prec {
			v.unexpected('+', "+ (one or more of preceding character)", "the preceding character must not be special character")
		}
		v.prec = false
	case '*':
		v.prec = false
	case '[':
		if v.scan.Peek() == ']' {
			v.unexpected(']', "content of character match []", "character match must not be empty")
		}
	Loop:
		for {
			c = v.scan.Next()
			switch c {
			case ']':
				break Loop
			case scanner.EOF:
				v.unexpected(c, "end of character match []", "missing ]")
				return false
			default:
				if v.scan.Peek() != '-' {
					// in case of single character
					break
				}
				s := c
				c = v.scan.Next() // eat -
				switch v.scan.Peek() {
				case ']':
					v.unexpected(c, "character range in []", "end of range is missing")
					c = v.scan.Next() // eat ]
					break Loop
				case scanner.EOF:
					// do nothing
				default:
					c = v.scan.Next() // eat end of range
					if s > c {
						why := fmt.Sprintf("start of range %q (%d) is larger than end of range %q (%d)", s, s, c, c)
						v.unexpected(c, "character range in []", why)
					}
				}
			}
		}
		v.prec = true
	case '\n':
		v.unexpected('\n', "", "newline cannot be contained")
		v.prec = true
	case ' ', '\t', '~', '^', ':':
		if v.isRef {
			v.invalidRefChar(c, "ref name cannot contain spaces, ~, ^, :, [, ?, *")
		}
		v.prec = true
	default:
		v.prec = true
	}

	if v.scan.Peek() == scanner.EOF {
		if v.isRef && (c == '/' || c == '.') {
			v.invalidRefChar(c, "ref name must not end with / and .")
		}
		return false
	}

	return true
}

func (v *globValidator) validate(pat string) {
	v.init(pat)

	if pat == "" {
		v.error("glob pattern cannot be empty")
		return
	}
	if pat == "!" {
		v.unexpected('!', "! at first character (negate pattern)", "at least one character must follow !")
		return
	}

	if v.isRef && v.scan.Peek() == '/' {
		v.invalidRefChar('/', "ref name must not start with /")
		v.scan.Next()
	}

	for v.validateNext() {
	}
}

func validateGlob(pat string, isRef bool) []InvalidGlobPattern {
	v := globValidator{}
	v.isRef = isRef
	v.validate(pat)
	return v.errs
}

// ValidateRefGlob checks a given input as glob pattern for Git ref names. It returns list of
// errors found by the validation. See the following URL for more details of the sytnax:
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
func ValidateRefGlob(pat string) []InvalidGlobPattern {
	return validateGlob(pat, true)
}

// ValidatePathGlob checks a given input as glob pattern for file paths. It returns list of
// errors found by the validation. See the following URL for more details of the sytnax:
// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
func ValidatePathGlob(pat string) []InvalidGlobPattern {
	return validateGlob(pat, false)
}
