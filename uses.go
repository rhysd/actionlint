package actionlint

import "strings"

// selfRepositoryUsesPrefix marks the "self-repository" form of `uses:`. It resolves against the
// repository running the workflow at the exact commit running it, so it carries no `@ref` and is
// accepted everywhere the workspace-relative `./` form is: workflow steps, composite action steps,
// nested composition, and reusable workflow calls.
// https://github.blog/changelog/2026-07-30-reference-same-repository-actions-with-self-repository-syntax/
const selfRepositoryUsesPrefix = "$/"

// canonLocalUsesSpec converts a `uses:` value naming an action or a reusable workflow in the
// workflow's own repository into its canonical `./{path}` form, and reports whether the value was
// such a reference at all.
//
// Callers normalise rather than test for either prefix because the metadata caches are keyed by
// spec. LocalReusableWorkflowCache.WriteWorkflowCallEvent writes its keys in the `./` form, so a
// `$/` caller looking the unnormalised spec up finds nothing that was written ahead of it.
func canonLocalUsesSpec(spec string) (string, bool) {
	if strings.HasPrefix(spec, "./") {
		return spec, true
	}
	if p, ok := strings.CutPrefix(spec, selfRepositoryUsesPrefix); ok {
		return "./" + p, true
	}
	return "", false
}
