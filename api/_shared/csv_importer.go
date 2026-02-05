package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// CreateGuestsTable creates the guests table if it doesn't exist
func (db *Database) CreateGuestsTable() error {
	if !db.IsConfigured() {
		return fmt.Errorf("database not configured")
	}

	log.Println("Ensuring guests table exists...")

	// SQL to create the table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS guests (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT NOT NULL,
			address TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_guests_name ON guests(name);
		CREATE INDEX IF NOT EXISTS idx_guests_address ON guests(address);
	`

	// Execute via Supabase SQL endpoint
	url := fmt.Sprintf("%s/rest/v1/rpc/exec_sql", db.url)

	payload := map[string]string{
		"query": createTableSQL,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SQL: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := db.client.Do(req)
	if err != nil {
		log.Printf("⚠️  Could not create table via SQL endpoint: %v", err)
		log.Printf("⚠️  Table creation skipped - ensure table exists manually")
		return nil // Don't fail import if table creation fails
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("⚠️  Table creation returned status %d", resp.StatusCode)
		log.Printf("⚠️  Assuming table already exists")
	} else {
		log.Println("✓ Guests table ready")
	}

	return nil
}

// ImportGuestsToDatabase imports guests from CSV to Supabase
func (db *Database) ImportGuestsToDatabase(guests []Guest) error {
	if !db.IsConfigured() {
		return fmt.Errorf("database not configured - cannot import")
	}

	// Try to create table first
	if err := db.CreateGuestsTable(); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Printf("Starting import process for %d guests...", len(guests))

	// Get existing guests from database
	existingGuests, err := db.LoadGuests()
	if err != nil {
		log.Printf("⚠️  Could not load existing guests: %v", err)
		log.Println("   Proceeding with import (may cause duplicates)")
		existingGuests = []Guest{}
	}

	// Create a map of existing guest names for quick lookup
	existingNames := make(map[string]bool)
	for _, guest := range existingGuests {
		existingNames[guest.Name] = true
	}

	// Filter out guests that already exist
	var newGuests []GuestRecord
	skippedCount := 0
	for _, guest := range guests {
		if existingNames[guest.Name] {
			skippedCount++
			continue
		}
		// Always include all fields to ensure consistency
		address := guest.Address
		if address == "" {
			address = "" // Ensure empty string, not null
		}
		newGuests = append(newGuests, GuestRecord{
			Name:    guest.Name,
			Address: address,
		})
	}

	log.Printf("Found %d existing guests, %d new guests to import", len(existingGuests), len(newGuests))

	if len(newGuests) == 0 {
		log.Println("✓ No new guests to import")
		return nil
	}

	log.Printf("Importing %d new guests to Supabase...", len(newGuests))

	// Use newGuests instead of records
	records := newGuests

	// Import in batches to avoid overwhelming the API
	batchSize := 50
	successCount := 0
	errorCount := 0

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]

		jsonData, err := json.Marshal(batch)
		if err != nil {
			log.Printf("Error marshaling batch %d: %v", i/batchSize+1, err)
			errorCount += len(batch)
			continue
		}

		url := fmt.Sprintf("%s/rest/v1/guests", db.url)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request for batch %d: %v", i/batchSize+1, err)
			errorCount += len(batch)
			continue
		}

		req.Header.Set("apikey", db.apiKey)
		req.Header.Set("Authorization", "Bearer "+db.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")

		resp, err := db.client.Do(req)
		if err != nil {
			log.Printf("❌ Error importing batch %d: %v", i/batchSize+1, err)
			errorCount += len(batch)
			continue
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			// Read response body for error details
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("❌ Batch %d returned status %d", i/batchSize+1, resp.StatusCode)
			log.Printf("   Response body: %s", string(bodyBytes))
			log.Printf("   First guest in batch: %s at %s", batch[0].Name, batch[0].Address)
			errorCount += len(batch)
		} else {
			resp.Body.Close()
			successCount += len(batch)
			log.Printf("✓ Imported batch %d (%d guests)", i/batchSize+1, len(batch))
		}
	}

	log.Printf("Import complete: %d successful, %d errors", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("import completed with %d errors", errorCount)
	}

	// Clear cache to force reload
	guestListMutex.Lock()
	guestListCache = []Guest{}
	guestListMutex.Unlock()

	return nil
}

// ImportCSVToDatabase loads a CSV file and imports it to the database
func ImportCSVToDatabase(csvFilePath string, db *Database) error {
	// Load guests from CSV
	guests, err := LoadGuestsFromCSV(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %v", err)
	}

	log.Printf("Loaded %d guests from %s", len(guests), csvFilePath)

	// Import to database
	if err := db.ImportGuestsToDatabase(guests); err != nil {
		return fmt.Errorf("failed to import to database: %v", err)
	}

	return nil
}
