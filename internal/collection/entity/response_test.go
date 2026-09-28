package entity

import (
	"encoding/json"
	"testing"
)

func TestCollectionListResponseJSONIncludesTotalCount(t *testing.T) {
	data, err := json.Marshal(CollectionListResponse{
		Collections: []CollectionListItemResponse{},
		TotalCount:  42,
	})
	if err != nil {
		t.Fatalf("marshal collection list response: %v", err)
	}

	got := string(data)
	want := `{"collections":[],"total_count":42}`
	if got != want {
		t.Fatalf("unexpected JSON: got %s, want %s", got, want)
	}
}

func TestCollectionListResponseJSONIncludesZeroTotalCount(t *testing.T) {
	data, err := json.Marshal(CollectionListResponse{
		Collections: []CollectionListItemResponse{},
	})
	if err != nil {
		t.Fatalf("marshal empty collection list response: %v", err)
	}

	got := string(data)
	want := `{"collections":[],"total_count":0}`
	if got != want {
		t.Fatalf("unexpected JSON: got %s, want %s", got, want)
	}
}
