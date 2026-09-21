package event_schema

import "cdk-infrastructure/database/models"

type SQLSchema struct {
	ClubID    int     `db:"club_id"`
	IsOwner   bool    `db:"is_owner"`
	ObjectKey *string `db:"object_key"`
	models.Event
	Description *string `db:"description"`
}

type ResponseSchema struct {
	models.Event
	Description        *string `json:"description"`
	models.EventOwners `json:"owners"`
}
