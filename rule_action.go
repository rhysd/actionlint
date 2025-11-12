package actionlint

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// BrandingColors is a set of colors allowed at branding.color in action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingcolor
var BrandingColors = map[string]struct{}{
	"white":     {},
	"black":     {},
	"yellow":    {},
	"blue":      {},
	"green":     {},
	"orange":    {},
	"red":       {},
	"purple":    {},
	"gray-dark": {},
}

// BrandingIcons is a set of icon names allowed at branding.icon in action.yaml.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingicon
var BrandingIcons = map[string]struct{}{
	"activity":           {},
	"airplay":            {},
	"alert-circle":       {},
	"alert-octagon":      {},
	"alert-triangle":     {},
	"align-center":       {},
	"align-justify":      {},
	"align-left":         {},
	"align-right":        {},
	"anchor":             {},
	"aperture":           {},
	"archive":            {},
	"arrow-down-circle":  {},
	"arrow-down-left":    {},
	"arrow-down-right":   {},
	"arrow-down":         {},
	"arrow-left-circle":  {},
	"arrow-left":         {},
	"arrow-right-circle": {},
	"arrow-right":        {},
	"arrow-up-circle":    {},
	"arrow-up-left":      {},
	"arrow-up-right":     {},
	"arrow-up":           {},
	"at-sign":            {},
	"award":              {},
	"bar-chart-2":        {},
	"bar-chart":          {},
	"battery-charging":   {},
	"battery":            {},
	"bell-off":           {},
	"bell":               {},
	"bluetooth":          {},
	"bold":               {},
	"book-open":          {},
	"book":               {},
	"bookmark":           {},
	"box":                {},
	"briefcase":          {},
	"calendar":           {},
	"camera-off":         {},
	"camera":             {},
	"cast":               {},
	"check-circle":       {},
	"check-square":       {},
	"check":              {},
	"chevron-down":       {},
	"chevron-left":       {},
	"chevron-right":      {},
	"chevron-up":         {},
	"chevrons-down":      {},
	"chevrons-left":      {},
	"chevrons-right":     {},
	"chevrons-up":        {},
	"circle":             {},
	"clipboard":          {},
	"clock":              {},
	"cloud-drizzle":      {},
	"cloud-lightning":    {},
	"cloud-off":          {},
	"cloud-rain":         {},
	"cloud-snow":         {},
	"cloud":              {},
	"code":               {},
	"command":            {},
	"compass":            {},
	"copy":               {},
	"corner-down-left":   {},
	"corner-down-right":  {},
	"corner-left-down":   {},
	"corner-left-up":     {},
	"corner-right-down":  {},
	"corner-right-up":    {},
	"corner-up-left":     {},
	"corner-up-right":    {},
	"cpu":                {},
	"credit-card":        {},
	"crop":               {},
	"crosshair":          {},
	"database":           {},
	"delete":             {},
	"disc":               {},
	"dollar-sign":        {},
	"download-cloud":     {},
	"download":           {},
	"droplet":            {},
	"edit-2":             {},
	"edit-3":             {},
	"edit":               {},
	"external-link":      {},
	"eye-off":            {},
	"eye":                {},
	"fast-forward":       {},
	"feather":            {},
	"file-minus":         {},
	"file-plus":          {},
	"file-text":          {},
	"file":               {},
	"film":               {},
	"filter":             {},
	"flag":               {},
	"folder-minus":       {},
	"folder-plus":        {},
	"folder":             {},
	"gift":               {},
	"git-branch":         {},
	"git-commit":         {},
	"git-merge":          {},
	"git-pull-request":   {},
	"globe":              {},
	"grid":               {},
	"hard-drive":         {},
	"hash":               {},
	"headphones":         {},
	"heart":              {},
	"help-circle":        {},
	"home":               {},
	"image":              {},
	"inbox":              {},
	"info":               {},
	"italic":             {},
	"layers":             {},
	"layout":             {},
	"life-buoy":          {},
	"link-2":             {},
	"link":               {},
	"list":               {},
	"loader":             {},
	"lock":               {},
	"log-in":             {},
	"log-out":            {},
	"mail":               {},
	"map-pin":            {},
	"map":                {},
	"maximize-2":         {},
	"maximize":           {},
	"menu":               {},
	"message-circle":     {},
	"message-square":     {},
	"mic-off":            {},
	"mic":                {},
	"minimize-2":         {},
	"minimize":           {},
	"minus-circle":       {},
	"minus-square":       {},
	"minus":              {},
	"monitor":            {},
	"moon":               {},
	"more-horizontal":    {},
	"more-vertical":      {},
	"move":               {},
	"music":              {},
	"navigation-2":       {},
	"navigation":         {},
	"octagon":            {},
	"package":            {},
	"paperclip":          {},
	"pause-circle":       {},
	"pause":              {},
	"percent":            {},
	"phone-call":         {},
	"phone-forwarded":    {},
	"phone-incoming":     {},
	"phone-missed":       {},
	"phone-off":          {},
	"phone-outgoing":     {},
	"phone":              {},
	"pie-chart":          {},
	"play-circle":        {},
	"play":               {},
	"plus-circle":        {},
	"plus-square":        {},
	"plus":               {},
	"pocket":             {},
	"power":              {},
	"printer":            {},
	"radio":              {},
	"refresh-ccw":        {},
	"refresh-cw":         {},
	"repeat":             {},
	"rewind":             {},
	"rotate-ccw":         {},
	"rotate-cw":          {},
	"rss":                {},
	"save":               {},
	"scissors":           {},
	"search":             {},
	"send":               {},
	"server":             {},
	"settings":           {},
	"share-2":            {},
	"share":              {},
	"shield-off":         {},
	"shield":             {},
	"shopping-bag":       {},
	"shopping-cart":      {},
	"shuffle":            {},
	"sidebar":            {},
	"skip-back":          {},
	"skip-forward":       {},
	"slash":              {},
	"sliders":            {},
	"smartphone":         {},
	"speaker":            {},
	"square":             {},
	"star":               {},
	"stop-circle":        {},
	"sun":                {},
	"sunrise":            {},
	"sunset":             {},
	"table":              {},
	"tablet":             {},
	"tag":                {},
	"target":             {},
	"terminal":           {},
	"thermometer":        {},
	"thumbs-down":        {},
	"thumbs-up":          {},
	"toggle-left":        {},
	"toggle-right":       {},
	"trash-2":            {},
	"trash":              {},
	"trending-down":      {},
	"trending-up":        {},
	"triangle":           {},
	"truck":              {},
	"tv":                 {},
	"type":               {},
	"umbrella":           {},
	"underline":          {},
	"unlock":             {},
	"upload-cloud":       {},
	"upload":             {},
	"user-check":         {},
	"user-minus":         {},
	"user-plus":          {},
	"user-x":             {},
	"user":               {},
	"users":              {},
	"video-off":          {},
	"video":              {},
	"voicemail":          {},
	"volume-1":           {},
	"volume-2":           {},
	"volume-x":           {},
	"volume":             {},
	"watch":              {},
	"wifi-off":           {},
	"wifi":               {},
	"wind":               {},
	"x-circle":           {},
	"x-square":           {},
	"x":                  {},
	"zap-off":            {},
	"zap":                {},
	"zoom-in":            {},
	"zoom-out":           {},
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsimage
func isImageOnDockerRegistry(image string) bool {
	return strings.HasPrefix(image, "docker://") ||
		strings.HasPrefix(image, "gcr.io/") ||
		strings.HasPrefix(image, "pkg.dev/") ||
		strings.HasPrefix(image, "ghcr.io/") ||
		strings.HasPrefix(image, "docker.io/")
}

// RuleAction is a rule to check running action in steps of jobs.
// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses
type RuleAction struct {
	RuleBase
	cache *LocalActionsCache
}

// NewRuleAction creates new RuleAction instance.
func NewRuleAction(cache *LocalActionsCache) *RuleAction {
	return &RuleAction{
		RuleBase: RuleBase{
			name: "action",
			desc: "Checks for popular actions released on GitHub, local actions, and action calls at \"uses:\"",
		},
		cache: cache,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleAction) VisitStep(n *Step) error {
	e, ok := n.Exec.(*ExecAction)
	if !ok || e.Uses == nil {
		return nil
	}

	if e.Uses.ContainsExpression() {
		// Cannot parse specification made with interpolation. Give up
		return nil
	}

	spec := e.Uses.Value

	if strings.HasPrefix(spec, "./") {
		// Relative to repository root
		rule.checkLocalAction(spec, e)
		return nil
	}

	if strings.HasPrefix(spec, "docker://") {
		rule.checkDockerAction(spec, e)
		return nil
	}

	rule.checkRepoAction(spec, e)
	return nil
}

// Parse {owner}/{repo}@{ref} or {owner}/{repo}/{path}@{ref}
func (rule *RuleAction) checkRepoAction(spec string, exec *ExecAction) {
	s := spec
	idx := strings.IndexRune(s, '@')
	if idx == -1 {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "ref is missing")
		return
	}
	ref := s[idx+1:]
	s = s[:idx] // remove {ref}

	idx = strings.IndexRune(s, '/')
	if idx == -1 {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "owner is missing")
		return
	}

	owner := s[:idx]
	s = s[idx+1:] // eat {owner}

	repo := s
	if idx := strings.IndexRune(s, '/'); idx >= 0 {
		repo = s[:idx]
	}

	if owner == "" || repo == "" || ref == "" {
		rule.invalidActionFormat(exec.Uses.Pos, spec, "owner and repo and ref should not be empty")
	}

	meta, ok := PopularActions[spec]
	if !ok {
		if _, ok := OutdatedPopularActionSpecs[spec]; ok {
			rule.Errorf(exec.Uses.Pos, "the runner of %q action is too old to run on GitHub Actions. update the action's version to fix this issue", spec)
			return
		}
		rule.Debug("This action is not found in popular actions data set: %s", spec)
		return
	}
	if meta.SkipInputs {
		rule.Debug("This action skips to check inputs: %s", spec)
		return
	}

	rule.checkAction(meta, exec, func(m *ActionMetadata) string {
		return strconv.Quote(spec)
	})
}

func (rule *RuleAction) invalidActionFormat(pos *Pos, spec string, why string) {
	rule.Errorf(pos, "specifying action %q in invalid format because %s. available formats are \"{owner}/{repo}@{ref}\" or \"{owner}/{repo}/{path}@{ref}\"", spec, why)
}

func (rule *RuleAction) missingRunsProp(pos *Pos, prop, ty, action, path string) {
	rule.Errorf(pos, `%q is required in "runs" section because %q is a %s action. the action is defined at %q`, prop, action, ty, path)
}

func (rule *RuleAction) checkInvalidRunsProps(pos *Pos, r *ActionMetadataRuns, ty, action, path string, props []string) {
	for _, prop := range props {
		invalid := prop == "main" && r.Main != "" ||
			prop == "pre" && r.Pre != "" ||
			prop == "pre-if" && r.PreIf != "" ||
			prop == "post" && r.Post != "" ||
			prop == "post-if" && r.PostIf != "" ||
			prop == "steps" && len(r.Steps) > 0 ||
			prop == "image" && r.Image != "" ||
			prop == "pre-entrypoint" && r.PreEntrypoint != "" ||
			prop == "entrypoint" && r.Entrypoint != "" ||
			prop == "post-entrypoint" && r.PostEntrypoint != "" ||
			prop == "args" && r.Args != nil ||
			prop == "env" && r.Env != nil

		if invalid {
			rule.Errorf(pos, `%q is not allowed in "runs" section because %q is a %s action. the action is defined at %q`, prop, action, ty, path)
		}
	}
}

func (rule *RuleAction) checkRunsFileExists(file, dir, prop, name string, pos *Pos) {
	f := filepath.FromSlash(file)
	if f == "" {
		return
	}
	p := filepath.Join(dir, f)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		rule.Errorf(pos, `file %q does not exist in %q. it is specified at %q key in "runs" section in %q action`, f, dir, prop, name)
	}
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-docker-container-actions
func (rule *RuleAction) checkLocalDockerActionRuns(r *ActionMetadataRuns, dir, name string, pos *Pos) {
	if r.Image == "" {
		rule.missingRunsProp(pos, "image", "Docker", name, dir)
	} else if !isImageOnDockerRegistry(r.Image) {
		rule.checkRunsFileExists(r.Image, dir, "image", name, pos)
		if filepath.Base(filepath.FromSlash(r.Image)) != "Dockerfile" {
			rule.Errorf(pos, `the local file %q referenced from "image" key must be named "Dockerfile" in %q action. the action is defined at %q`, r.Image, name, dir)
		}
	}
	rule.checkRunsFileExists(r.PreEntrypoint, dir, "pre-entrypoint", name, pos)
	rule.checkRunsFileExists(r.Entrypoint, dir, "entrypoint", name, pos)
	rule.checkRunsFileExists(r.PostEntrypoint, dir, "post-entrypoint", name, pos)
	rule.checkInvalidRunsProps(pos, r, "Docker", name, dir, []string{"main", "pre", "pre-if", "post", "post-if", "steps"})
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-composite-actions
func (rule *RuleAction) checkLocalCompositeActionRuns(r *ActionMetadataRuns, dir, name string, pos *Pos) {
	if r.Steps == nil {
		rule.missingRunsProp(pos, "steps", "Composite", name, dir)
	}
	rule.checkInvalidRunsProps(pos, r, "Composite", name, dir, []string{"main", "pre", "pre-if", "post", "post-if", "image", "pre-entrypoint", "entrypoint", "post-entrypoint", "args", "env"})
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-javascript-actions
func (rule *RuleAction) checkLocalJavaScriptActionRuns(r *ActionMetadataRuns, dir, name string, pos *Pos) {
	if r.Main == "" {
		rule.missingRunsProp(pos, "main", "JavaScript", name, dir)
	} else {
		rule.checkRunsFileExists(r.Main, dir, "main", name, pos)
	}

	rule.checkRunsFileExists(r.Pre, dir, "pre", name, pos)
	if r.Pre == "" && r.PreIf != "" {
		rule.Errorf(pos, `"pre" is required when "pre-if" is specified in "runs" section in %q action. the action is defined at %q`, name, dir)
	}

	rule.checkRunsFileExists(r.Post, dir, "post", name, pos)
	if r.Post == "" && r.PostIf != "" {
		rule.Errorf(pos, `"post" is required when "post-if" is specified in "runs" section in %q action. the action is defined at %q`, name, dir)
	}

	rule.checkInvalidRunsProps(pos, r, "JavaScript", name, dir, []string{"steps", "image", "pre-entrypoint", "entrypoint", "post-entrypoint", "args", "env"})
}

func (rule *RuleAction) checkLocalActionInputs(meta *ActionMetadata, pos *Pos) {
	for _, i := range meta.Inputs {
		if i.Deprecated && i.DeprecationMessage == "" {
			rule.Errorf(
				pos,
				"input %q is deprecated but \"deprecationMessage\" is empty in metadata of %q action at %q",
				i.Name,
				meta.Name,
				meta.Path(),
			)
		}
	}
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs
func (rule *RuleAction) checkLocalActionRuns(meta *ActionMetadata, pos *Pos) {
	switch r := &meta.Runs; r.Using {
	case "":
		rule.Errorf(pos, `"runs.using" is missing in local action %q defined at %q`, meta.Name, meta.Dir())
	case "docker":
		rule.checkLocalDockerActionRuns(r, meta.Dir(), meta.Name, pos)
	case "composite":
		rule.checkLocalCompositeActionRuns(r, meta.Dir(), meta.Name, pos)
	case "node20", "node24":
		rule.checkLocalJavaScriptActionRuns(r, meta.Dir(), meta.Name, pos)
	default:
		rule.Errorf(pos, `invalid runner name %q at runs.using in %q action defined at %q. valid runners are "composite", "docker", "node20", and "node24". see https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs`, r.Using, meta.Name, meta.Dir())

		// Probably invalid version of Node.js runner. Assume it is JavaScript action to find as many errors as possible
		if strings.HasPrefix(r.Using, "node") {
			rule.checkLocalJavaScriptActionRuns(r, meta.Dir(), meta.Name, pos)
		}
	}
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-using-the-github-packages-container-registry
func (rule *RuleAction) checkDockerAction(uri string, exec *ExecAction) {
	tag := ""
	tagExists := false
	if idx := strings.IndexRune(uri[len("docker://"):], ':'); idx != -1 {
		idx += len("docker://")
		if idx < len(uri) {
			tag = uri[idx+1:]
			uri = uri[:idx]
			tagExists = true
		}
	}

	if _, err := url.Parse(uri); err != nil {
		rule.Errorf(
			exec.Uses.Pos,
			"URI for Docker container %q is invalid: %s (tag=%s)",
			uri,
			err.Error(),
			tag,
		)
	}

	if tagExists && tag == "" {
		rule.Errorf(exec.Uses.Pos, "tag of Docker action should not be empty: %q", uri)
	}
}

// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
func (rule *RuleAction) checkLocalActionMetadata(meta *ActionMetadata, action *ExecAction) {
	if meta.Name == "" {
		rule.Errorf(action.Uses.Pos, "name is required in action metadata %q", meta.Path())
	}
	if meta.Description == "" {
		rule.Errorf(action.Uses.Pos, "description is required in metadata of %q action at %q", meta.Name, meta.Path())
	}
	if meta.Branding.Icon != "" {
		if _, ok := BrandingIcons[strings.ToLower(meta.Branding.Icon)]; !ok {
			rule.Errorf(
				action.Uses.Pos,
				"incorrect icon name %q at branding.icon in metadata of %q action at %q. see the official document to know the exhaustive list of supported icons: https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingicon",
				meta.Branding.Icon,
				meta.Name,
				meta.Path(),
			)
		}
	}
	if meta.Branding.Color != "" {
		if _, ok := BrandingColors[strings.ToLower(meta.Branding.Color)]; !ok {
			rule.Errorf(
				action.Uses.Pos,
				"incorrect color %q at branding.icon in metadata of %q action at %q. see the official document to know the exhaustive list of supported colors: https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingcolor",
				meta.Branding.Color,
				meta.Name,
				meta.Path(),
			)
		}
	}
	rule.checkLocalActionInputs(meta, action.Uses.Pos)
	rule.checkLocalActionRuns(meta, action.Uses.Pos)
}

// https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#example-using-action-in-the-same-repository-as-the-workflow
func (rule *RuleAction) checkLocalAction(spec string, action *ExecAction) {
	meta, cached, err := rule.cache.FindMetadata(spec)
	if err != nil {
		rule.Error(action.Uses.Pos, err.Error())
		return
	}
	if meta == nil {
		return
	}

	if !cached {
		rule.Debug("Checking metadata of %s action %q at %q", meta.Runs, meta.Name, spec)
		rule.checkLocalActionMetadata(meta, action)
	}

	rule.checkAction(meta, action, func(m *ActionMetadata) string {
		return fmt.Sprintf("%q defined at %q", m.Name, spec)
	})
}

var reNewlineWithIndent = regexp.MustCompile(`\r?\n\s*`)

func (rule *RuleAction) checkAction(meta *ActionMetadata, exec *ExecAction, describe func(*ActionMetadata) string) {
	// Check specified inputs are defined in action's inputs spec
	for id, i := range exec.Inputs {
		m, ok := meta.Inputs[id]
		if !ok {
			ns := make([]string, 0, len(meta.Inputs))
			for _, i := range meta.Inputs {
				ns = append(ns, i.Name)
			}
			rule.Errorf(
				i.Name.Pos,
				"input %q is not defined in action %s. available inputs are %s",
				i.Name.Value,
				describe(meta),
				sortedQuotes(ns),
			)
		} else if m.Deprecated {
			msg := fmt.Sprintf(
				"avoid using deprecated input %q in action %s",
				i.Name.Value,
				describe(meta),
			)
			d := reNewlineWithIndent.ReplaceAllString(strings.TrimRight(m.DeprecationMessage, ". "), " ")
			if d != "" {
				msg += ": " + d
			}
			rule.Error(i.Name.Pos, msg)
		}
	}

	// Check mandatory inputs are specified
	for id, i := range meta.Inputs {
		if i.Required {
			if _, ok := exec.Inputs[id]; !ok {
				ns := make([]string, 0, len(meta.Inputs))
				for _, i := range meta.Inputs {
					if i.Required {
						ns = append(ns, i.Name)
					}
				}
				rule.Errorf(
					exec.Uses.Pos,
					"missing input %q which is required by action %s. all required inputs are %s",
					i.Name,
					describe(meta),
					sortedQuotes(ns),
				)
			}
		}
	}
}
