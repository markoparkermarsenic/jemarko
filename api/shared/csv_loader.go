package shared

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

// LoadGuestsFromCSV loads guests from a CSV file
func LoadGuestsFromCSV(filename string) ([]Guest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []Guest{}, nil
	}

	// Skip header row and parse guests
	var guests []Guest
	idCounter := 1

	for i, record := range records {
		// Skip header row
		if i == 0 {
			continue
		}

		// Ensure we have at least name and address columns
		if len(record) < 2 {
			continue
		}

		name := strings.TrimSpace(record[0])
		address := strings.TrimSpace(record[1])

		// Skip empty names
		if name == "" || name == "Name" {
			continue
		}

		// Keep N/A addresses as-is (don't normalize)

		// Add dietary restrictions if available (column 5)
		dietary := ""
		if len(record) > 5 {
			dietary = strings.TrimSpace(record[5])
		}

		guest := Guest{
			ID:      string(rune(idCounter)),
			Name:    name,
			Address: address,
		}

		// Store dietary info in a way that can be used later if needed
		_ = dietary // For future use

		guests = append(guests, guest)
		idCounter++
	}

	log.Printf("✓ Loaded %d guests from CSV file: %s", len(guests), filename)
	return guests, nil
}
