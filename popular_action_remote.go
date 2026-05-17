package actionlint

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.yaml.in/yaml/v4"
)

const minOutdatedNodeRunnerVersion = 20

var (
	reGitCommitRef = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

	// rawGitHubusercontentBase is the prefix for fetching action metadata. Tests may override it.
	rawGitHubusercontentBase = "https://raw.githubusercontent.com"

	remoteActionMetadataCache sync.Map
)

func isGitCommitRef(ref string) bool {
	return reGitCommitRef.MatchString(ref)
}

func popularActionSlug(owner, repo, actionPath string) string {
	if actionPath == "" {
		return owner + "/" + repo
	}
	return owner + "/" + repo + actionPath
}

func popularActionHasOutdatedRelease(slug string) bool {
	prefix := slug + "@"
	for spec := range OutdatedPopularActionSpecs {
		if strings.HasPrefix(spec, prefix) {
			return true
		}
	}
	return false
}

func actionUsesOutdatedNodeRunner(using string) bool {
	if !strings.HasPrefix(using, "node") {
		return false
	}
	v, err := strconv.ParseUint(using[len("node"):], 10, 8)
	return err == nil && v < minOutdatedNodeRunnerVersion
}

func rawActionMetadataURL(owner, repo, actionPath, ref string) string {
	return fmt.Sprintf("%s/%s/%s/%s%s/action.yml", rawGitHubusercontentBase, owner, repo, ref, actionPath)
}

func fetchRemoteActionMetadata(owner, repo, actionPath, ref string) (*ActionMetadata, error) {
	key := owner + "/" + repo + actionPath + "@" + ref
	if cached, ok := remoteActionMetadataCache.Load(key); ok {
		if cached == nil {
			return nil, nil
		}
		return cached.(*ActionMetadata), nil
	}

	url := rawActionMetadataURL(owner, repo, actionPath, ref)
	c := &http.Client{Timeout: 15 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		url = strings.TrimSuffix(url, ".yml") + ".yaml"
		res, err = c.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err := fmt.Errorf("request was not successful %s: %s", url, res.Status)
		remoteActionMetadataCache.Store(key, nil)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var meta ActionMetadata
	if err := yaml.Unmarshal(body, &meta); err != nil {
		return nil, err
	}

	remoteActionMetadataCache.Store(key, &meta)
	return &meta, nil
}
