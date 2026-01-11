package gh

import (
	"testing"
)

func TestSearchResult_GenericTypes(t *testing.T) {
	// Verify the generic SearchResult works with custom types
	type TestItem struct {
		ID   int
		Name string
	}

	result := SearchResult[TestItem]{
		TotalCount: 2,
		Items: []TestItem{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
		},
	}

	if result.TotalCount != 2 {
		t.Errorf("SearchResult.TotalCount = %d, want %d", result.TotalCount, 2)
	}

	if len(result.Items) != 2 {
		t.Errorf("len(SearchResult.Items) = %d, want %d", len(result.Items), 2)
	}

	if result.Items[0].ID != 1 || result.Items[0].Name != "first" {
		t.Errorf("SearchResult.Items[0] = %+v, want {ID: 1, Name: first}", result.Items[0])
	}
}
