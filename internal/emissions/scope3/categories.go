package scope3

// Category represents a Scope 3 category placeholder.
type Category struct {
	ID   string
	Name string
}

// Categories returns placeholder categories.
func Categories() []Category {
	return []Category{
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
}
