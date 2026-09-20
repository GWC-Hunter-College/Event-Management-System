package models

type EventImage struct {
	EventID string `db:"fk_event_id"   json:"eventId"`
	ImageID string `db:"fk_image_id"   json:"imageId"`
}
