package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/ingestion"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: inspect_activities <org-id>")
		os.Exit(1)
	}
	org := os.Args[1]
	dsn := "postgresql://offgridflow:changeme@localhost:5432/offgridflow?sslmode=disable"
	database, err := db.Connect(context.Background(), db.Config{DSN: dsn})
	if err != nil {
		fmt.Printf("connect error: %v\n", err)
		os.Exit(2)
	}
	defer database.Close()

	store := ingestion.NewPostgresActivityStore(database.DB)
	acts, err := store.ListByOrgAndSource(context.Background(), org, "utility_bill")
	if err != nil {
		fmt.Printf("query error: %v\n", err)
		os.Exit(3)
	}
	fmt.Printf("found %d activities for org %s\n", len(acts), org)
	for i, a := range acts {
		fmt.Printf("%d: id=%s source=%s unit=%s qty=%v org=%s period_start=%v period_end=%v\n", i, a.ID, a.Source, a.Unit, a.Quantity, a.OrgID, a.PeriodStart, a.PeriodEnd)
	}
}
