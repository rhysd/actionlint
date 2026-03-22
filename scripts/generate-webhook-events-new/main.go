package main

// This is a script to generate a Go source that contains all activity types of all webhook events.
// Run the following command from the root of this repository to apply manually.
// This script is usually run via `go generate`.
// ```
// go run ./generate-webhook-events-new .\input.html all_webhooks.go
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

	htmlpkg "golang.org/x/net/html"
)

var theURL = "https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows"
var dbg = log.New(io.Discard, "", log.LstdFlags)

// Parse the activity types of each webhook event. The keys of the map are names of the webhook events
// like "pull_request", and the values are arrays of names of their activity types.
// The HTML input is assumed to be fetched from the following page.
// https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#pull_request
func parseWebhookActivityTypes(html []byte) (map[string][]string, error) {
	doc, err := htmlpkg.Parse(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML: %w", err)
	}
	dbg.Printf("Parsed HTML document (%d bytes)\n", len(html))

	body := findNode(doc, func(n *htmlpkg.Node) bool {
		return n.Type == htmlpkg.ElementNode &&
			n.Data == "div" &&
			attr(n, "data-search") == "article-body"
	})
	if body == nil {
		return nil, errors.New("article body was not found in HTML")
	}
	dbg.Println(`Found article body container "div[data-search=article-body]"`)

	markdown := findNode(body, func(n *htmlpkg.Node) bool {
		return n.Type == htmlpkg.ElementNode &&
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
		if n.Type != htmlpkg.ElementNode {
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

		types, err := parseWebhookTypesTable(currentHook, n)
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
func findNode(root *htmlpkg.Node, pred func(*htmlpkg.Node) bool) *htmlpkg.Node {
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

func attr(n *htmlpkg.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *htmlpkg.Node, want string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == want {
			return true
		}
	}
	return false
}

func textContent(n *htmlpkg.Node) string {
	var b strings.Builder
	var visit func(*htmlpkg.Node)
	visit = func(n *htmlpkg.Node) {
		if n.Type == htmlpkg.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
	return strings.TrimSpace(b.String())
}

func eventNameOfHeading(h *htmlpkg.Node) string {
	if id := attr(h, "id"); id != "" {
		return id
	}
	name := textContent(h)
	dbg.Printf("Using heading text as hook name because id was missing: %q\n", name)
	return name
}

func elementChildren(n *htmlpkg.Node, tag string) []*htmlpkg.Node {
	var nodes []*htmlpkg.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == htmlpkg.ElementNode && c.Data == tag {
			nodes = append(nodes, c)
		}
	}
	return nodes
}

func parseWebhookTypesTable(hook string, table *htmlpkg.Node) ([]string, error) {
	label := attr(table, "aria-labelledby")
	dbg.Printf("Table: %q\n", label)
	if label != hook {
		return nil, fmt.Errorf("table aria-labelledby %q did not match the expected hook id %q", label, hook)
	}

	thead := firstElementChildByTag(table, "thead")
	if thead == nil {
		return nil, errors.New("thead element was missing")
	}
	tr := firstElementChildByTag(thead, "tr")
	if tr == nil {
		return nil, errors.New("header row was missing")
	}
	headers := elementChildren(tr, "th")
	if len(headers) < 2 {
		return nil, fmt.Errorf("table header had too few columns: got %d, want at least 2", len(headers))
	}
	h0 := textContent(headers[0])
	h1 := textContent(headers[1])
	if h0 != "Webhook event payload" {
		return nil, fmt.Errorf("unexpected first table header %q, want %q", h0, "Webhook event payload")
	}
	if h1 != "Activity types" {
		return nil, fmt.Errorf("unexpected second table header %q, want %q", h1, "Activity types")
	}
	dbg.Println(`  Found table header for "Webhook event payload"`)

	tbody := firstElementChildByTag(table, "tbody")
	if tbody == nil {
		return nil, errors.New("tbody element was missing")
	}
	row := firstElementChildByTag(tbody, "tr")
	if row == nil {
		return nil, errors.New("table row was missing")
	}
	dbg.Println("  Found the first table row")
	cells := elementChildren(row, "td")
	if len(cells) < 2 {
		return nil, errors.New("table did not have at least two columns")
	}

	name := textContent(cells[0])
	dbg.Printf("  First column text: %q\n", name)

	types := codeTexts(cells[1])
	if len(types) > 0 {
		dbg.Printf("  Activity types from code elements: %v\n", types)
		return types, nil
	}

	t := textContent(cells[1])
	if t == "" || strings.EqualFold(t, "Not applicable") {
		dbg.Printf("  Activity types cell treated as empty set: %q\n", t)
		return []string{}, nil
	}
	if strings.EqualFold(t, "Custom") {
		dbg.Printf("  Activity types cell treated as custom types: %q\n", t)
		return nil, nil
	}

	return nil, fmt.Errorf("activity types cell did not contain code elements nor 'Not applicable': %q", t)
}

func firstElementChildByTag(n *htmlpkg.Node, tag string) *htmlpkg.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == htmlpkg.ElementNode && c.Data == tag {
			return c
		}
	}
	return nil
}

func codeTexts(n *htmlpkg.Node) []string {
	var texts []string
	var visit func(*htmlpkg.Node)
	visit = func(n *htmlpkg.Node) {
		if n.Type == htmlpkg.ElementNode && n.Data == "code" {
			t := textContent(n)
			if t != "" {
				texts = append(texts, t)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
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

	var html []byte
	var err error
	if len(args) == 2 {
		html, err = os.ReadFile(args[0])
	} else {
		html, err = fetch(srcURL)
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
		n := args[len(args)-1]
		f, err := os.Create(n)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
		dst = n
	}

	m, err := parseWebhookActivityTypes(html)
	if err != nil {
		return err
	}

	if err := write(m, out); err != nil {
		return err
	}

	dbg.Println("Wrote output to", dst)
	dbg.Println("Done generate-webhook-events script successfully")

	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, theURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
