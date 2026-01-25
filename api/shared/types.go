package shared

// Guest represents a guest on the invite list
type Guest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VerifyNameRequest represents a name verification request
type VerifyNameRequest struct {
	Name string `json:"name"`
}

// VerifyNameResponse represents a name verification response
type VerifyNameResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// RSVPRequest represents an RSVP submission
type RSVPRequest struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	IsAttending     bool     `json:"isAttending"`
	AttendingGuests []string `json:"attendingGuests,omitempty"`
	Diet            string   `json:"diet,omitempty"`
}

// RSVPResponse represents an RSVP submission response
type RSVPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
