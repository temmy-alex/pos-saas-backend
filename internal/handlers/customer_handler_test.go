package handlers

import (
	"testing"

	"pos-saas-backend/internal/requests"
)

func TestNormalizeCustomerFilter(t *testing.T) {
	tests := []struct {
		name       string
		input      requests.CustomerFilterRequest
		wantPage   int
		wantLimit  int
		wantSearch string
	}{
		{
			name:       "uses defaults and trims search",
			input:      requests.CustomerFilterRequest{Search: "  alex  "},
			wantPage:   1,
			wantLimit:  10,
			wantSearch: "alex",
		},
		{
			name:      "supports mobile per page parameter",
			input:     requests.CustomerFilterRequest{Page: 2, PerPage: 25},
			wantPage:  2,
			wantLimit: 25,
		},
		{
			name:      "caps page size",
			input:     requests.CustomerFilterRequest{Page: 1, Limit: 500},
			wantPage:  1,
			wantLimit: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizeCustomerFilter(test.input)
			if got.Page != test.wantPage {
				t.Fatalf("page = %d, want %d", got.Page, test.wantPage)
			}
			if got.Limit != test.wantLimit {
				t.Fatalf("limit = %d, want %d", got.Limit, test.wantLimit)
			}
			if got.Search != test.wantSearch {
				t.Fatalf("search = %q, want %q", got.Search, test.wantSearch)
			}
		})
	}
}
