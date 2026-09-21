package authentication_utils

import (
	"context"
	"errors"

	"cdk-infrastructure/utils/query_client"
)

var (
	ErrBadEvent = errors.New("invalid event id")
)

// AuthorizeStudentEvent returns true if the student is eboard/owner
// of any club associated with the given event.
func AuthorizeStudentEvent(ctx context.Context, qc *query_client.QueryClient, sub string, eventId int) (bool, error) {
	if sub == "" {
		return false, ErrNoSub
	}
	if eventId <= 0 {
		return false, ErrBadEvent
	}

	var authorized bool
	q := query_client.NewQuery("authorization/IS_student_authorized_event.sql", eventId, sub)
	if err := qc.Get(&authorized, q); err != nil {
		return false, err
	}
	return authorized, nil
}
