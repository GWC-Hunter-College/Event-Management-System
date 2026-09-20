package models

type Club struct {
	ID           int     `json:"id,omitempty" db:"id"`
	Name         string  `json:"name" db:"name" validate:"required"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty" db:"thumbnail_url"`
}

type ClubRole string

const (
	RoleMember ClubRole = "member"
	RoleEboard ClubRole = "eboard"
	RoleOwner  ClubRole = "owner"
)

type ClubWithRole struct {
	Club
	ClubRole string `json:"role" db:"role"` // "owner" | "eboard" | "member"
}

type ClubDetailed struct {
	Club
	WebsiteURL  *string `db:"website_url" json:"website_url"`
	Description *string `db:"description" json:"description"`
}
