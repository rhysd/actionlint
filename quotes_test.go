package actionlint

import (
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestQuotesSortedQuotes(t *testing.T) {
	testCases := [][]string{
		{},
		{"foo"},
		{"foo", "bar", "piyo"},
		{"\n", "\t"},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ss := tc[:]
			sort.Strings(ss)
			qs := make([]string, 0, len(ss))
			for _, s := range tc {
				qs = append(qs, strconv.Quote(s))
			}
			want := strings.Join(qs, ", ")
			have := sortedQuotes(tc)
			if want != have {
				t.Errorf("want: %s\nhave: %s", want, have)
			}
		})
	}
}

func TestQuotesQuotesRunes(t *testing.T) {
	testCases := [][]rune{
		{},
		{'a'},
		{'a', 'b', 'c'},
		{'\n', '\t', '\\'},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			qs := make([]string, 0, len(tc))
			for _, r := range tc {
				qs = append(qs, strconv.QuoteRune(r))
			}
			want := strings.Join(qs, ", ")
			qb := quotesBuilder{}
			for _, r := range tc {
				qb.appendRune(r)
			}
			have := qb.build()
			if want != have {
				t.Errorf("want: %s\nhave: %s", want, have)
			}
		})
	}
}

func TestQuotesQuotesAll(t *testing.T) {
	testCases := [][][]string{
		{{}},
		{{"foo"}},
		{{"foo", "bar"}},
		{{"foo"}, {"bar"}},
		{{}, {"foo"}, {"bar", "piyo"}},
		{{"\n", "\t"}, {"\v"}, {}},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			qs := []string{}
			for _, ss := range tc {
				for _, s := range ss {
					qs = append(qs, strconv.Quote(s))
				}
			}
			want := strings.Join(qs, ", ")
			have := quotesAll(tc...)
			if want != have {
				t.Errorf("want: %s\nhave: %s", want, have)
			}
		})
	}
}
