package shared

import (
	"strings"
)

// NormalizeString converts a string to lowercase and trims whitespace for comparison
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// IsValidEmail performs basic email validation
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// IsGuestInList checks if a guest name exists in the guest list
func IsGuestInList(name string, guestList []Guest) bool {
	normalizedName := NormalizeString(name)
	for _, guest := range guestList {
		if NormalizeString(guest.Name) == normalizedName {
			return true
		}
	}
	return false
}

// FindGuest searches for a guest by name in the guest list
func FindGuest(name string, guestList []Guest) *Guest {
	normalizedName := NormalizeString(name)
	for i := range guestList {
		if NormalizeString(guestList[i].Name) == normalizedName {
			return &guestList[i]
		}
	}
	return nil
}
