package club_authorization

import (
	"context"
	"errors"

	"cdk-infrastructure/utils/query_client"
)

var (
	ErrNoSub   = errors.New("empty sub")
	ErrBadClub = errors.New("invalid club id")
)

// AuthorizerStudent returns true if the student is eboard/owner for the given club.
func AuthorizeStudentClub(ctx context.Context, qc *query_client.QueryClient, sub string, clubId int) (bool, error) {
	if sub == "" {
		return false, ErrNoSub
	}
	if clubId <= 0 {
		return false, ErrBadClub
	}

	var authorized bool // if your qc.Get doesn't scan into bool, use int and compare == 1
	q := query_client.NewQuery("authorization/IS_student_authorized_club.sql", sub, clubId)
	if err := qc.Get(&authorized, q); err != nil {
		return false, err
	}
	return authorized, nil
}
