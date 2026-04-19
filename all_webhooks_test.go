package actionlint

import "testing"

func TestGeneratedAllWebhooks(t *testing.T) {
	if len(AllWebhookTypes) == 0 {
		t.Fatal("AllWebhookTypes is empty")
	}

	for name, types := range AllWebhookTypes {
		if name == "" {
			t.Errorf("Name is empty (types=%v)", types)
			continue
		}

		seen := map[string]struct{}{}
		for _, ty := range types {
			if _, ok := seen[ty]; ok {
				t.Errorf("type %q duplicates in webhook %q: %v", ty, name, types)
			} else {
				seen[ty] = struct{}{}
			}
		}
	}
}
