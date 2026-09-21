package models

type Event struct {
	EventID     int     `json:"id" db:"id" validate:"omitempty"`
	AuthorID    string  `json:"authorId" db:"fk_author_id"`
	ThumbnailID *string `json:"thumbnailId,omitempty" db:"fk_thumbnail_id" validate:"omitempty,uuid"`
	Title       string  `json:"title" db:"title" validate:"required"`
	Location    string  `json:"location" db:"location" validate:"required"`
	RsvpLink    *string `json:"rsvpLink" db:"rsvp_link"`
	Status      string  `json:"status" db:"status" validate:"omitempty"`
	StartDate   string  `json:"startDate" db:"start_date" validate:"required"`
	EndDate     string  `json:"endDate" db:"end_date" validate:"required"`
	Timezone    string  `json:"timezone" db:"timezone" validate:"required"`
	CreatedAt   string  `json:"createdAt" db:"created_at"`
	UpdatedAt   string  `json:"updatedAt" db:"updated_at"`
	DeletedAt   *string `json:"deletedAt,omitempty" db:"deleted_at"`
}

type FullEvent struct {
	Event
	Description *string `json:"description,omitempty"`
}

type EventDescription struct {
	EventID     int     `json:"eventId" db:"fk_event_id"`
	Description *string `json:"description,omitempty" db:"description"`
}

type EventOwners struct {
	Owner      Club   `json:"owner"`
	Associates []Club `json:"associates"`
}
