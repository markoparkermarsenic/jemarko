package shared

// Guest represents a guest on the invite list
type Guest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
	Avatar  string `json:"avatar,omitempty"`
	Message string `json:"message,omitempty"`
}

// VerifyNameRequest represents a name verification request
type VerifyNameRequest struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// FamilyMember represents a family member in the verification response
type FamilyMember struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VerifyNameResponse represents a name verification response
type VerifyNameResponse struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message,omitempty"`
	FamilyMembers []FamilyMember `json:"familyMembers,omitempty"`
}

// RSVPRequest represents an RSVP submission
type RSVPRequest struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	IsAttending     bool     `json:"isAttending"`
	AttendingGuests []string `json:"attendingGuests,omitempty"`
	Diet            string   `json:"diet,omitempty"`
	Verified        bool     `json:"verified,omitempty"`
}

// RSVPResponse represents an RSVP submission response
type RSVPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AvatarSelection represents a single guest's avatar selection
type AvatarSelection struct {
	GuestName string `json:"guestName"`
	Avatar    string `json:"avatar"`
	Message   string `json:"message"`
}

// SaveAvatarsRequest represents a request to save avatar selections
type SaveAvatarsRequest struct {
	Email   string            `json:"email"`
	Avatars []AvatarSelection `json:"avatars"`
}

// SaveAvatarsResponse represents the response after saving avatars
type SaveAvatarsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GuestAvatar represents a guest with their avatar
type GuestAvatar struct {
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
}

// GetAvatarsResponse represents the response for getting avatars
type GetAvatarsResponse struct {
	Success bool          `json:"success"`
	Avatars []GuestAvatar `json:"avatars"`
}
