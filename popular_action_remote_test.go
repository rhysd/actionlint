package actionlint

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestIsGitCommitRef(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ref  string
		want bool
	}{
		{"2b250bc32ad02700b996b496c14ac8c2840a2991", true},
		{"v2.1.8", false},
		{"main", false},
		{"abc", false},
	}
	for _, tc := range tests {
		if got := isGitCommitRef(tc.ref); got != tc.want {
			t.Errorf("isGitCommitRef(%q) = %v, want %v", tc.ref, got, tc.want)
		}
	}
}

func TestActionUsesOutdatedNodeRunner(t *testing.T) {
	t.Parallel()
	if !actionUsesOutdatedNodeRunner("node12") {
		t.Fatal("node12 should be outdated")
	}
	if actionUsesOutdatedNodeRunner("node20") {
		t.Fatal("node20 should not be outdated")
	}
}

func TestFetchRemoteActionMetadataOutdated(t *testing.T) {
	const body = "runs:\n  using: node12\n  main: dist/index.js\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/actions/cache/deadbeef/action.yml" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	oldBase := rawGitHubusercontentBase
	rawGitHubusercontentBase = srv.URL
	t.Cleanup(func() { rawGitHubusercontentBase = oldBase })
	remoteActionMetadataCache = sync.Map{}

	meta, err := fetchRemoteActionMetadata("actions", "cache", "", "deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if meta == nil || !actionUsesOutdatedNodeRunner(meta.Runs.Using) {
		t.Fatalf("expected outdated node12 metadata, got %#v", meta)
	}
}
