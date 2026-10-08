package handlers

import (
	"testing"

	"pos-saas-backend/internal/requests"
)

func TestNormalizeStockOpnameFilter(t *testing.T) {
	t.Run("normalizes pagination and text", func(t *testing.T) {
		filter, err := normalizeStockOpnameFilter(requests.StockOpnameFilterRequest{
			Status:  " COMPLETED ",
			Search:  "  SO-123 ",
			PerPage: 25,
		})
		if err != nil {
			t.Fatalf("normalizeStockOpnameFilter returned error: %v", err)
		}
		if filter.Status != "completed" || filter.Search != "SO-123" {
			t.Fatalf("unexpected normalized text: status=%q search=%q", filter.Status, filter.Search)
		}
		if filter.Page != 1 || filter.Limit != 25 {
			t.Fatalf("unexpected pagination: page=%d limit=%d", filter.Page, filter.Limit)
		}
	})

	t.Run("rejects unknown status", func(t *testing.T) {
		_, err := normalizeStockOpnameFilter(requests.StockOpnameFilterRequest{Status: "approved"})
		if err == nil {
			t.Fatal("normalizeStockOpnameFilter should reject an unknown status")
		}
	})

	t.Run("caps page size", func(t *testing.T) {
		filter, err := normalizeStockOpnameFilter(requests.StockOpnameFilterRequest{Limit: 500})
		if err != nil {
			t.Fatalf("normalizeStockOpnameFilter returned error: %v", err)
		}
		if filter.Limit != 100 {
			t.Fatalf("limit = %d, want 100", filter.Limit)
		}
	})
}
