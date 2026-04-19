package actionlint

import (
	"regexp"
	"slices"
	"testing"
)

func TestWorkflowKeyAvailability(t *testing.T) {
	seenCtx := map[string]struct{}{}
	for _, key := range allWorkflowKeys {
		t.Run(key, func(t *testing.T) {
			ctx, sp := WorkflowKeyAvailability(key)
			if ctx == nil || sp == nil {
				t.Error("workflow key has not availability info:", key)
			}
			if len(ctx) == 0 {
				t.Error("no context is available for key", key)
			}
			for _, c := range ctx {
				ks, ok := AllContexts[c]
				if !ok {
					t.Errorf("Context %q is not included in AllContexts %v", c, AllContexts)
				} else if !slices.Contains(ks, key) {
					t.Errorf("Workflow key %q is not included in workflow keys of AllContexts %v", key, ks)
				}
				seenCtx[c] = struct{}{}
			}

			r := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
			for _, c := range ctx {
				if !r.MatchString(c) {
					t.Errorf("context %q does not match to pattern %s", c, r)
				}
			}
			for _, s := range sp {
				if !r.MatchString(s) {
					t.Errorf("context %q does not match to pattern %s", s, r)
				}
				ks, ok := SpecialFunctionNames[s]
				if !ok {
					t.Errorf("special function %q is not registered in SpecialFunctionNames: %v", s, SpecialFunctionNames)
				}
				ok = false
				for _, k := range ks {
					if k == key {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("Key %q is not in candidates of special function %q: %v", key, s, ks)
				}
			}
		})
	}
	if len(AllContexts) != len(seenCtx) {
		t.Errorf(
			"AllContexts has %d elems but %d contexts were found. Found contexts are %v and AllContexts is %v",
			len(AllContexts),
			len(seenCtx),
			seenCtx,
			AllContexts,
		)
	}

	ctx, sp := WorkflowKeyAvailability("unknown.workflow.key")
	if len(ctx) != 0 {
		t.Error("some context was returned", ctx)
	}
	if len(sp) != 0 {
		t.Error("some special function name was returned", sp)
	}
}

func TestSpecialFunctionNames(t *testing.T) {
	if len(SpecialFunctionNames) == 0 {
		t.Error("No special function is registered in SpecialFunctionNames")
	}

	r := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	for f, ws := range SpecialFunctionNames {
		if len(ws) == 0 {
			t.Errorf("No workflow key is available for special function %q", f)
		}
		if !r.MatchString(f) {
			t.Errorf("Special function name does not match to pattern %s", r)
		}
	}
}
