package shared

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"
)

// normaliseHeader lowercases, trims, and strips trailing "?" so headers like
// "Address " or "Ceremony?" match reliably regardless of formatting.
func normaliseHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.TrimSuffix(h, "?")
	return strings.TrimSpace(h)
}

// parseBool interprets common truthy/falsy CSV values case-insensitively:
// true/false, yes/no, y/n, 1/0. Anything unrecognised defaults to false.
func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "y", "1":
		return true
	case "false", "no", "n", "0", "":
		return false
	default:
		if b, err := strconv.ParseBool(strings.TrimSpace(s)); err == nil {
			return b
		}
		return false
	}
}

// LoadGuestsFromCSV loads guests from a CSV file. Columns are located by
// header name (case-insensitive, trims whitespace/trailing "?"), so the
// CSV can have columns in any order as long as headers include at least
// "Name". "Address" and "Ceremony" are optional.
func LoadGuestsFromCSV(filename string) ([]Guest, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Allow rows with a variable number of fields (defensive against trailing commas)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []Guest{}, nil
	}

	// Build a header → column index map
	header := records[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[normaliseHeader(col)] = i
	}

	nameIdx, hasName := colIndex["name"]
	if !hasName {
		log.Printf("⚠️  CSV missing required 'Name' column header: %v", header)
		return []Guest{}, nil
	}
	addressIdx, hasAddress := colIndex["address"]
	ceremonyIdx, hasCeremony := colIndex["ceremony"]

	getField := func(record []string, idx int) string {
		if idx < 0 || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	var guests []Guest
	idCounter := 1

	for i, record := range records {
		if i == 0 {
			continue // skip header row
		}

		name := getField(record, nameIdx)
		if name == "" {
			continue
		}

		address := ""
		if hasAddress {
			address = getField(record, addressIdx)
		}

		ceremony := false
		if hasCeremony {
			ceremony = parseBool(getField(record, ceremonyIdx))
		}

		guests = append(guests, Guest{
			ID:       strconv.Itoa(idCounter),
			Name:     name,
			Address:  address,
			Ceremony: ceremony,
		})
		idCounter++
	}

	log.Printf("✓ Loaded %d guests from CSV file: %s", len(guests), filename)
	return guests, nil
}
