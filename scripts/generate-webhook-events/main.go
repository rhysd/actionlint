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

	"github.com/PuerkitoBio/goquery"
)

var theURL = "https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows"
var dbg = log.New(io.Discard, "", log.LstdFlags)

// Parse the activity types of each webhook event. The keys of the map are names of the webhook events
// like "pull_request", and the values are arrays of names of their activity types.
// The HTML input is assumed to be fetched from the following page.
// https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#pull_request
func parse(src []byte) (map[string][]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(src)))
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML: %w", err)
	}
	dbg.Printf("Parsed HTML document (%d bytes)\n", len(src))

	body := doc.Find(`div[data-search="article-body"]`).First()
	if body.Length() == 0 {
		return nil, errors.New("article body was not found in HTML")
	}
	dbg.Println(`Found article body container "div[data-search=article-body]"`)

	markdown := body.Find("div.markdown-body").First()
	if markdown.Length() == 0 {
		return nil, errors.New("markdown body was not found in HTML")
	}
	dbg.Println(`Found markdown body container "div.markdown-body"`)

	parsed := map[string][]string{}
	about := markdown.ChildrenFiltered(`h2#about-events-that-trigger-workflows`).First()
	if about.Length() == 0 {
		return nil, errors.New(`"About events that trigger workflows" heading was missing`)
	}
	dbg.Println(`Found "About events that trigger workflows" heading`)

	headings := about.NextAllFiltered("h2")
	for i := range headings.Length() {
		h := headings.Eq(i)
		hook := eventNameOfHeading(h)
		dbg.Printf("Found new hook %q\n", hook)

		table := h.NextUntil("h2").Filter("table").First()
		if table.Length() == 0 {
			dbg.Printf("Skipping hook %q because no table was found before the next h2\n", hook)
			continue
		}

		label, _ := table.Attr("aria-labelledby")
		dbg.Printf("Trying table for hook %q (aria-labelledby=%q)\n", hook, label)

		types, err := parseTable(hook, table)
		if err != nil {
			return nil, fmt.Errorf("could not parse webhook types table for %q: %w", hook, err)
		}

		parsed[hook] = types
		dbg.Printf("Found webhook table: %q %v\n", hook, types)
	}
	if len(parsed) == 0 {
		return nil, errors.New("no webhook table was found in given HTML source")
	}
	dbg.Printf("Parsed %d webhook tables\n", len(parsed))

	return parsed, nil
}

func eventNameOfHeading(h *goquery.Selection) string {
	if id, ok := h.Attr("id"); ok && id != "" {
		return id
	}
	name := strings.TrimSpace(h.Text())
	dbg.Printf("Using heading text as hook name because id was missing: %q\n", name)
	return name
}

func parseTable(hook string, table *goquery.Selection) ([]string, error) {
	label, _ := table.Attr("aria-labelledby")
	dbg.Printf("Table: %q\n", label)
	if label != hook {
		return nil, fmt.Errorf("table aria-labelledby %q did not match the expected hook id %q", label, hook)
	}

	thead := table.ChildrenFiltered("thead").First()
	if thead.Length() == 0 {
		return nil, errors.New("thead element was missing")
	}
	tr := thead.ChildrenFiltered("tr").First()
	if tr.Length() == 0 {
		return nil, errors.New("header row was missing")
	}
	headers := tr.ChildrenFiltered("th")
	if headers.Length() < 2 {
		return nil, fmt.Errorf("table header had too few columns: got %d, want at least 2", headers.Length())
	}
	h0 := strings.TrimSpace(headers.Eq(0).Text())
	h1 := strings.TrimSpace(headers.Eq(1).Text())
	if h0 != "Webhook event payload" {
		return nil, fmt.Errorf("unexpected first table header %q, want %q", h0, "Webhook event payload")
	}
	if h1 != "Activity types" {
		return nil, fmt.Errorf("unexpected second table header %q, want %q", h1, "Activity types")
	}
	dbg.Println(`  Found table header for "Webhook event payload"`)

	tbody := table.ChildrenFiltered("tbody").First()
	if tbody.Length() == 0 {
		return nil, errors.New("tbody element was missing")
	}
	row := tbody.ChildrenFiltered("tr").First()
	if row.Length() == 0 {
		return nil, errors.New("table row was missing")
	}
	dbg.Println("  Found the first table row")
	cells := row.ChildrenFiltered("td")
	if cells.Length() < 2 {
		return nil, errors.New("table did not have at least two columns")
	}

	name := strings.TrimSpace(cells.Eq(0).Text())
	dbg.Printf("  First column text: %q\n", name)

	types := codeTexts(cells.Eq(1))
	if len(types) > 0 {
		dbg.Printf("  Activity types from code elements: %v\n", types)
		return types, nil
	}

	t := strings.TrimSpace(cells.Eq(1).Text())
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

func codeTexts(n *goquery.Selection) []string {
	codes := n.Find("code")
	texts := make([]string, 0, codes.Length())
	codes.Each(func(_ int, s *goquery.Selection) {
		if t := strings.TrimSpace(s.Text()); t != "" {
			texts = append(texts, t)
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
		n := args[len(args)-1]
		f, err := os.Create(n)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
		dst = n
	}

	m, err := parse(src)
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
