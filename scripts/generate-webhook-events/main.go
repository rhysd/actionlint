package main

// This is a script to generate a Go source that contains all activity types of all webhook events.
// Run the following command from the root of this repository to apply manually.
// This script is usually run via `go generate`.
// ```
// go run ./scripts/generate-webhook-events input.html all_webhooks.go
// ```

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

const theURL = "https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows"

var dbg = log.New(io.Discard, "", log.LstdFlags)

// Parse the activity types of each webhook event. The keys of the map are names of the webhook events
// like "pull_request", and the values are arrays of names of their activity types.
// The HTML input is assumed to be fetched from the following page.
// https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#pull_request
func parse(src []byte) (map[string][]string, error) {
	doc, err := html.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML: %w", err)
	}
	dbg.Printf("Parsed HTML document (%d bytes)\n", len(src))

	body := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			attr(n, "data-search") == "article-body"
	})
	if body == nil {
		return nil, errors.New("article body was not found in HTML")
	}
	dbg.Println(`Found article body container "div[data-search=article-body]"`)

	markdown := findNode(body, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "markdown-body")
	})
	if markdown == nil {
		return nil, errors.New("markdown body was not found in HTML")
	}
	dbg.Println(`Found markdown body container "div.markdown-body"`)

	parsed := map[string][]string{}
	sawAbout := false
	currentHook := ""

	for n := markdown.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode {
			continue
		}

		if n.Data == "h2" {
			id := attr(n, "id")
			if id == "about-events-that-trigger-workflows" {
				sawAbout = true
				currentHook = ""
				dbg.Println(`Found "About events that trigger workflows" heading`)
				continue
			}
			if !sawAbout {
				dbg.Printf("Skipping h2 %q because the about heading has not been seen yet\n", id)
				continue
			}

			currentHook = eventNameOfHeading(n)
			dbg.Printf("Found new hook %q\n", currentHook)
			continue
		}

		if !sawAbout || currentHook == "" || n.Data != "table" {
			continue
		}

		dbg.Printf("Trying table for hook %q (aria-labelledby=%q)\n", currentHook, attr(n, "aria-labelledby"))

		types, err := parseTable(currentHook, n)
		if err != nil {
			return nil, fmt.Errorf("could not parse webhook types table for %q: %w", currentHook, err)
		}

		parsed[currentHook] = types
		dbg.Printf("Found webhook table: %q %v\n", currentHook, types)
		currentHook = ""
	}

	if !sawAbout {
		return nil, errors.New(`"About events that trigger workflows" heading was missing`)
	}
	if len(parsed) == 0 {
		return nil, errors.New("no webhook table was found in given HTML source")
	}
	dbg.Printf("Parsed %d webhook tables\n", len(parsed))

	return parsed, nil
}

func findNode(root *html.Node, pred func(*html.Node) bool) *html.Node {
	if pred(root) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if n := findNode(c, pred); n != nil {
			return n
		}
	}
	return nil
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, want string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == want {
			return true
		}
	}
	return false
}

func walk(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, visit)
	}
}

func text(n *html.Node) string {
	var b strings.Builder
	walk(n, func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
	})
	return strings.TrimSpace(b.String())
}

func eventNameOfHeading(h *html.Node) string {
	if id := attr(h, "id"); id != "" {
		return id
	}
	name := text(h)
	dbg.Printf("Using heading text as hook name because id was missing: %q\n", name)
	return name
}

func children(n *html.Node, tag string) []*html.Node {
	var nodes []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			nodes = append(nodes, c)
		}
	}
	return nodes
}

func firstChildByTag(n *html.Node, tag string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return c
		}
	}
	return nil
}

func parseTable(hook string, table *html.Node) ([]string, error) {
	label := attr(table, "aria-labelledby")
	dbg.Printf("Table: %q\n", label)
	if label != hook {
		return nil, fmt.Errorf(`expected table[aria-labelledby] to be %q, got %q`, hook, label)
	}

	thead := firstChildByTag(table, "thead")
	if thead == nil {
		return nil, errors.New("missing thead element")
	}
	tr := firstChildByTag(thead, "tr")
	if tr == nil {
		return nil, errors.New("missing header row in thead")
	}
	headers := children(tr, "th")
	if len(headers) < 2 {
		return nil, fmt.Errorf("expected at least 2 header columns, got %d", len(headers))
	}
	h0 := text(headers[0])
	h1 := text(headers[1])
	if h0 != "Webhook event payload" {
		return nil, fmt.Errorf(`expected first header to be "Webhook event payload", got %q`, h0)
	}
	if h1 != "Activity types" {
		return nil, fmt.Errorf(`expected second header to be "Activity types", got %q`, h1)
	}
	dbg.Println(`  Found table header for "Webhook event payload"`)

	tbody := firstChildByTag(table, "tbody")
	if tbody == nil {
		return nil, errors.New("missing tbody element")
	}
	row := firstChildByTag(tbody, "tr")
	if row == nil {
		return nil, errors.New("missing first data row in tbody")
	}
	dbg.Println("  Found the first table row")
	cells := children(row, "td")
	if len(cells) < 2 {
		return nil, fmt.Errorf("expected at least 2 data columns, got %d", len(cells))
	}

	name := text(cells[0])
	dbg.Printf("  First column text: %q\n", name)

	types := code(cells[1])
	if len(types) > 0 {
		dbg.Printf("  Activity types from code elements: %v\n", types)
		return types, nil
	}

	t := text(cells[1])
	if t == "" || strings.EqualFold(t, "Not applicable") {
		dbg.Printf("  Activity types cell treated as empty set: %q\n", t)
		return []string{}, nil
	}
	if strings.EqualFold(t, "Custom") {
		dbg.Printf("  Activity types cell treated as custom types: %q\n", t)
		return nil, nil
	}

	return nil, fmt.Errorf(`unexpected activity types cell text %q; expected code elements, "Not applicable", or "Custom"`, t)
}

func code(n *html.Node) []string {
	var texts []string
	walk(n, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "code" {
			t := text(n)
			if t != "" {
				texts = append(texts, t)
			}
		}
	})
	return texts
}

func write(parsed map[string][]string, out io.Writer) error {
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, `// Code generated by actionlint/scripts/generate-webhook-events. DO NOT EDIT.

package actionlint

// AllWebhookTypes is a table of all webhooks with their types. This variable was generated by
// script at ./scripts/generate-webhook-events based on
// https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
// The value is nil when the activity types cannot be determined from the document. For example
// repository_dispatch event can contain arbitrary types that are customized by user.
var AllWebhookTypes = map[string][]string{`)

	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ts := parsed[k]
		if ts == nil {
			fmt.Fprintf(buf, "\t%q: nil,\n", k)
			continue
		}
		if len(ts) == 0 {
			fmt.Fprintf(buf, "\t%q: {},\n", k)
			continue
		}
		fmt.Fprintf(buf, "\t%q: {%q", k, ts[0])
		for _, t := range ts[1:] {
			fmt.Fprintf(buf, ", %q", t)
		}
		fmt.Fprintln(buf, "},")
	}
	fmt.Fprintln(buf, "}")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not format Go source: %w", err)
	}

	if _, err := out.Write(src); err != nil {
		return fmt.Errorf("could not write output: %w", err)
	}

	return nil
}

func fetch(url string) ([]byte, error) {
	var c http.Client

	dbg.Println("Fetching", url)

	res, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch %s: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || 300 <= res.StatusCode {
		return nil, fmt.Errorf("request was not successful for %s: %s", url, res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not fetch body for %s: %w", url, err)
	}

	dbg.Printf("Fetched %d bytes from %s", len(body), url)
	return body, nil
}

func run(args []string, stdout, dbgout io.Writer, srcURL string) error {
	dbg.SetOutput(dbgout)

	if len(args) > 2 {
		return errors.New("usage: generate-webhook-events [[srcfile] dstfile]")
	}

	dbg.Println("Start generate-webhook-events script")

	var src []byte
	var err error
	if len(args) == 2 {
		src, err = os.ReadFile(args[0])
	} else {
		src, err = fetch(srcURL)
	}
	if err != nil {
		return err
	}

	var out io.Writer
	var dst string
	if len(args) == 0 || args[len(args)-1] == "-" {
		out = stdout
		dst = "stdout"
	} else {
		dst = args[len(args)-1]
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}

	p, err := parse(src)
	if err != nil {
		return err
	}

	if err := write(p, out); err != nil {
		return err
	}

	dbg.Println("Wrote the output to", dst)
	dbg.Println("Done generate-webhook-events script successfully")

	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, theURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
