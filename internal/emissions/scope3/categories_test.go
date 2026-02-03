package scope3

import "testing"

func TestCategoriesCanonical(t *testing.T) {
	expected := []Category{
		{ID: "1", Name: "Purchased Goods and Services"},
		{ID: "2", Name: "Capital Goods"},
		{ID: "3", Name: "Fuel- and Energy-Related Activities"},
		{ID: "4", Name: "Upstream Transportation and Distribution"},
		{ID: "5", Name: "Waste Generated in Operations"},
		{ID: "6", Name: "Business Travel"},
		{ID: "7", Name: "Employee Commuting"},
		{ID: "8", Name: "Upstream Leased Assets"},
		{ID: "9", Name: "Downstream Transportation and Distribution"},
		{ID: "10", Name: "Processing of Sold Products"},
		{ID: "11", Name: "Use of Sold Products"},
		{ID: "12", Name: "End-of-Life Treatment of Sold Products"},
		{ID: "13", Name: "Downstream Leased Assets"},
		{ID: "14", Name: "Franchises"},
		{ID: "15", Name: "Investments"},
	}

	got := Categories()
	if len(got) != len(expected) {
		t.Fatalf("expected %d categories, got %d", len(expected), len(got))
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("category %d mismatch: expected %+v, got %+v", i, expected[i], got[i])
		}
	}
}

func TestCategoriesReturnsCopy(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 {
		t.Fatalf("expected categories, got none")
	}

	cats[0].Name = "mutated"
	again := Categories()
	if again[0].Name == "mutated" {
		t.Fatalf("expected Categories() to return a copy, got shared slice")
	}
}
