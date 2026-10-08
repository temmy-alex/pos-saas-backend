package handlers

import (
	"testing"
	"time"
)

func TestParseBusinessDate(t *testing.T) {
	t.Run("accepts ISO date", func(t *testing.T) {
		got, err := parseBusinessDate("2026-10-08")
		if err != nil {
			t.Fatalf("parseBusinessDate returned error: %v", err)
		}
		if got != "2026-10-08" {
			t.Fatalf("date = %q, want %q", got, "2026-10-08")
		}
	})

	t.Run("rejects non ISO date", func(t *testing.T) {
		if _, err := parseBusinessDate("08-10-2026"); err == nil {
			t.Fatal("parseBusinessDate should reject a non-ISO date")
		}
	})

	t.Run("defaults to Jakarta current date", func(t *testing.T) {
		got, err := parseBusinessDate("")
		if err != nil {
			t.Fatalf("parseBusinessDate returned error: %v", err)
		}
		location, locationErr := time.LoadLocation("Asia/Jakarta")
		if locationErr != nil {
			location = time.FixedZone("Asia/Jakarta", 7*60*60)
		}
		want := time.Now().In(location).Format("2006-01-02")
		if got != want {
			t.Fatalf("date = %q, want %q", got, want)
		}
	})
}
