package main

import (
"context"
"fmt"
"log"
"time"

"github.com/example/offgridflow/internal/ingestion/sources/sap"
)

// Example demonstrates how to use the SAP adapter to ingest energy and emissions data
func main() {
// Configure the SAP adapter
// In production, load these from environment variables
cfg := sap.Config{
BaseURL:      "https://api.sap.yourcompany.com",
ClientID:     "your-client-id",
ClientSecret: "your-client-secret",
Company:      "1000",              // SAP company code
Plant:        "US-TX-001",         // Optional: filter by plant
OrgID:        "org-yourcompany",   // OffGridFlow organization ID
StartDate:    time.Now().AddDate(0, -1, 0), // Last month
EndDate:      time.Now(),
}

// Create the adapter
adapter, err := sap.NewAdapter(cfg)
if err != nil {
log.Fatalf("Failed to create SAP adapter: %v", err)
}

// Ingest data from SAP
log.Println("Starting SAP data ingestion...")
activities, err := adapter.Ingest(context.Background())
if err != nil {
log.Fatalf("Failed to ingest SAP data: %v", err)
}

// Process the ingested activities
log.Printf("Successfully ingested %d activities from SAP\n", len(activities))

// Example: Categorize and summarize the data
energyActivities := 0
emissionsActivities := 0
totalEnergy := 0.0
totalEmissions := 0.0

for _, activity := range activities {
switch activity.Source {
case "sap_erp":
energyActivities++
if activity.Unit == "kWh" || activity.Unit == "MWh" {
totalEnergy += activity.Quantity
}

case "sap_sustainability":
emissionsActivities++
if activity.Unit == "kg" {
totalEmissions += activity.Quantity
}
}

// Print sample activity
if energyActivities+emissionsActivities <= 5 {
fmt.Printf("\nActivity %d:\n", energyActivities+emissionsActivities)
fmt.Printf("  Source: %s\n", activity.Source)
fmt.Printf("  Category: %s\n", activity.Category)
fmt.Printf("  Quantity: %.2f %s\n", activity.Quantity, activity.Unit)
fmt.Printf("  Period: %s to %s\n",
activity.PeriodStart.Format("2006-01-02"),
activity.PeriodEnd.Format("2006-01-02"))
fmt.Printf("  Location: %s\n", activity.Location)
if activity.MeterID != "" {
fmt.Printf("  Meter: %s\n", activity.MeterID)
}
if plant, ok := activity.Metadata["sap_plant"]; ok {
fmt.Printf("  Plant: %s\n", plant)
}
}
}

// Print summary
fmt.Println("\n=== Ingestion Summary ===")
fmt.Printf("Energy activities: %d\n", energyActivities)
fmt.Printf("Emissions activities: %d\n", emissionsActivities)
fmt.Printf("Total energy: %.2f kWh\n", totalEnergy)
fmt.Printf("Total emissions: %.2f kg CO2e\n", totalEmissions)

// Example: Filter activities by category
fmt.Println("\n=== Activities by Category ===")
categoryCount := make(map[string]int)
for _, activity := range activities {
categoryCount[activity.Category]++
}
for category, count := range categoryCount {
fmt.Printf("  %s: %d activities\n", category, count)
}

// Example: Activities by location
fmt.Println("\n=== Activities by Location ===")
locationCount := make(map[string]int)
for _, activity := range activities {
locationCount[activity.Location]++
}
for location, count := range locationCount {
fmt.Printf("  %s: %d activities\n", location, count)
}
}
