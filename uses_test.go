package actionlint

import "testing"

func TestCanonLocalUsesSpec(t *testing.T) {
	tests := []struct {
		what string
		spec string
		want string
		ok   bool
	}{
		{"workspace relative action", "./action", "./action", true},
		{"workspace relative nested path", "./.github/actions/my-action", "./.github/actions/my-action", true},
		{"workspace relative workflow", "./.github/workflows/ci.yml", "./.github/workflows/ci.yml", true},
		{"self repository action", "$/action", "./action", true},
		{"self repository nested path", "$/.github/actions/my-action", "./.github/actions/my-action", true},
		{"self repository workflow", "$/.github/workflows/ci.yml", "./.github/workflows/ci.yml", true},
		{"self repository with empty path", "$/", "./", true},
		{"repository action", "owner/repo@v1", "", false},
		{"repository action with path", "owner/repo/path@v1", "", false},
		{"docker action", "docker://alpine:3.18", "", false},
		{"dollar without slash", "$action", "", false},
		{"dollar alone", "$", "", false},
		{"parent relative path", "../action", "", false},
		{"empty", "", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.what, func(t *testing.T) {
			have, ok := canonLocalUsesSpec(tc.spec)
			if ok != tc.ok {
				t.Fatalf("wanted ok=%v but have ok=%v for spec %q", tc.ok, ok, tc.spec)
			}
			if have != tc.want {
				t.Fatalf("wanted %q but have %q for spec %q", tc.want, have, tc.spec)
			}
		})
	}
}

// Both spellings must land on the same cache key, which is the property the metadata caches depend
// on: LocalReusableWorkflowCache.WriteWorkflowCallEvent writes its keys in the "./" form, and a
// "$/" caller that looked up anything else would never find them.
func TestCanonLocalUsesSpecAgreesBetweenForms(t *testing.T) {
	for _, path := range []string{"action", ".github/actions/my-action", ".github/workflows/ci.yml"} {
		local, ok := canonLocalUsesSpec("./" + path)
		if !ok {
			t.Fatalf("%q was not recognised as a local spec", "./"+path)
		}
		self, ok := canonLocalUsesSpec("$/" + path)
		if !ok {
			t.Fatalf("%q was not recognised as a local spec", "$/"+path)
		}
		if local != self {
			t.Errorf("%q and %q canonicalise differently: %q vs %q", "./"+path, "$/"+path, local, self)
		}
	}
}
