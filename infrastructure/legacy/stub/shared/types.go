package shared

type EventStatus string

const (
	StatusDrafted  EventStatus = "drafted"
	StatusPosted   EventStatus = "posted"
	StatusArchived EventStatus = "archived"
)

type Event struct {
	ID       int         `json:"id"`
	Title    string      `json:"title"`
	Location string      `json:"location,omitempty"`
	RSVPLink string      `json:"rsvpLink,omitempty"`
	Status   EventStatus `json:"status"`
	StartISO string      `json:"startDate"`
	EndISO   string      `json:"endDate"`
	// Timezone string      `json:"timezone"`
}

type Events struct {
	Events []Event `json:"clubs"`
}

type EventDetailed struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Location    string      `json:"location,omitempty"`
	RSVPLink    string      `json:"rsvpLink,omitempty"`
	Status      EventStatus `json:"status"`
	StartISO    string      `json:"startDate"`
	EndISO      string      `json:"endDate"`
	Timezone    string      `json:"timezone,omitempty"`
	CreateISO   string      `json:"createDate"`
	UpdateISO   string      `json:"updateDate"`
	Description string      `json:"description,omitempty"`
}

type Image struct {
	ID        int    `json:"id"`
	Purpose   string `json:"purpose"`
	URL       string `json:"url"`
	CreateISO string `json:"createDate"`
}

type Images struct {
	Images []Image `json:"images"`
}

type EventDescription struct {
	Description string `json:"description"`
}

type EventOwners struct {
	Owner      Club   `json:"owner"`
	Associates []Club `json:"associates"`
}

type ClubRole string

const (
	RoleMember ClubRole = "member"
	RoleEboard ClubRole = "eboard"
	RoleOwner  ClubRole = "owner"
)

type Club struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Role         ClubRole `json:"role,omitempty"`
	ThumbnailURL string   `json:"thumbnailUrl,omitempty"`
}

type Clubs struct {
	Clubs []Club `json:"clubs"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
