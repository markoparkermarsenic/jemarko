package main

import (
	"flag"
	"log"
	"os"

	"utils/shared"
)

func main() {
	// Parse command line flags
	csvFile := flag.String("file", "../invite_list.csv", "Path to CSV file to import")
	flag.Parse()

	log.Printf("CSV Import Tool")
	log.Printf("===============")
	log.Printf("CSV File: %s", *csvFile)
	log.Printf("")

	// Check if file exists
	if _, err := os.Stat(*csvFile); os.IsNotExist(err) {
		log.Fatalf("❌ CSV file not found: %s", *csvFile)
	}

	// Create database connection
	db := shared.NewDatabase()

	if !db.IsConfigured() {
		log.Println("⚠️  Supabase not configured!")
		log.Println("Please set SUPABASE_URL and SUPABASE_API_KEY environment variables")
		log.Println("")
		log.Println("You can set them in api/.env file:")
		log.Println("  SUPABASE_URL=your-supabase-url")
		log.Println("  SUPABASE_API_KEY=your-api-key")
		log.Fatal("❌ Cannot import without database configuration")
	}

	log.Printf("✓ Database configured")
	log.Printf("")

	// Import CSV to database
	if err := shared.ImportCSVToDatabase(*csvFile, db); err != nil {
		log.Fatalf("❌ Import failed: %v", err)
	}

	log.Printf("")
	log.Printf("✅ Import completed successfully!")
}
