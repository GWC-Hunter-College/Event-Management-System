package authentication_utils

import (
	"context"
	"database/sql"
	"errors"

	"cdk-infrastructure/utils/query_client"
)

var ErrNoSub = errors.New("jwt has no sub")
var ErrNoAuthorizer = errors.New("Request context has no authorizer")

// RequireStudent ensures a student row exists for this sub.
// Return values (error):
//   - Returns ErrNoSub if the provided sub is empty
//   - Returns nil if the student already exists (no action needed)
//   - Returns nil if the student did not exist and was successfully inserted/upserted
//   - Returns an error if the DB existence check fails (not sql.ErrNoRows)
//   - Returns an error if the insert/upsert statement fails
//
// Flow:
// 1) If sub is empty -> return ErrNoSub
// 2) If EXISTS(sub) succeeds -> return nil
// 3) If DB error during EXISTS -> return error
// 4) If missing -> insert (with email if provided, else sub-only)
// 5) If insert fails -> return error; else return nil
func RequireStudent(ctx context.Context, qc *query_client.QueryClient, sub string, email string) error {
	if sub == "" {
		return ErrNoSub
	}

	// 1) Fast existence check by sub
	var exists int
	existsQ := query_client.NewQuery("students/EXISTS_student_by_sub.sql", sub)
	if err := qc.Get(&exists, existsQ); err != nil {
		// No row -> proceed to insert
		if errors.Is(err, sql.ErrNoRows) {
			// Missing: create it
			if email != "" {
				upsertQuery := query_client.NewQuery("students/UPSERT_student.sql", sub, email)
				_, execErr := qc.Exec(upsertQuery)
				return execErr // nil if success, error if DB insert fails
			}
			// No email: upsert sub-only (does not touch email column if already present)
			subUpsertQuery := query_client.NewQuery("students/UPSERT_student_sub_only.sql", sub)
			_, execErr := qc.Exec(subUpsertQuery)
			return execErr // nil if success, error if DB insert fails
		}
		// Real DB error on EXISTS
		return err
	}

	// Row exists -> nothing to do
	return nil
}
