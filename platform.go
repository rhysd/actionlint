package actionlint

import (
	"net/url"
	"os"
	"strings"
)

// isGitHubHostedPlatform reports whether actionlint should apply checks that only
// apply to github.com or GitHub Enterprise Cloud (*.ghe.com). When GITHUB_SERVER_URL
// points to GitHub Enterprise Server or another self-hosted instance, some checks are
// skipped because shipped actions and runner policies can differ (#509).
func isGitHubHostedPlatform() bool {
	raw := strings.TrimSpace(os.Getenv("GITHUB_SERVER_URL"))
	if raw == "" {
		return true
	}

	raw = strings.TrimSuffix(raw, "/")
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return true
	}

	host := strings.ToLower(u.Hostname())
	if host == "github.com" {
		return true
	}
	if strings.HasSuffix(host, ".ghe.com") {
		return true
	}

	return false
}
