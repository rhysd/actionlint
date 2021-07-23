package actionlint

import "testing"

func TestPopularActionsDataset(t *testing.T) {
	if len(PopularActions) == 0 {
		t.Fatal("popular actions data set is empty")
	}

	for n, meta := range PopularActions {
		if meta == nil {
			t.Fatalf("metadata for %s is nil", n)
		}
		if meta.Name == "" {
			t.Fatalf("action name for %s is empty", n)
		}
	}
}
