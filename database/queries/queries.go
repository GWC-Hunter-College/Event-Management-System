// Package queries loads the application SQL: the queries copied from the legacy
// query client and the queries written for the documented endpoints.
package queries

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed students/ensure/*.sql me/get/*.sql me/clubs/list/*.sql me/events/list/*.sql
//go:embed clubs/list/*.sql clubs/get/*.sql clubs/create/*.sql clubs/update/*.sql clubs/tags/replace/*.sql
//go:embed clubs/members/create/*.sql clubs/members/leave/*.sql clubs/members/list/*.sql clubs/members/update_role/*.sql
//go:embed clubs/events/list/*.sql clubs/events/drafts/*.sql clubs/thumbnails/confirm/*.sql
//go:embed clubs/verification/create/*.sql clubs/verification/delete/*.sql
//go:embed events/read/*.sql events/create/*.sql events/update/*.sql events/delete/*.sql events/restore/*.sql events/purge/*.sql
//go:embed events/images/list/*.sql events/images/confirm/*.sql events/thumbnails/confirm/*.sql
//go:embed images/create/*.sql images/get/*.sql images/release/*.sql images/delete/*.sql images/purge/*.sql
//go:embed authorization/clubs/can_manage/*.sql authorization/clubs/is_member/*.sql authorization/clubs/is_owner/*.sql
//go:embed authorization/events/can_manage/*.sql authorization/events/manages_owner_club/*.sql authorization/admins/is_admin/*.sql
//go:embed admins/list/*.sql admins/get/*.sql admins/create/*.sql admins/delete/*.sql
var files embed.FS

// Load returns the complete SQL file at a slash-separated path relative to this
// package, for example "students/ensure/UPSERT_student.sql". It preserves the
// original file bytes, including whitespace, without padding or truncation.
func Load(path string) (string, error) {
	return load(files, path)
}

func load(source fs.FS, path string) (string, error) {
	content, err := fs.ReadFile(source, path)
	if err != nil {
		return "", fmt.Errorf("load SQL %q: %w", path, err)
	}
	return string(content), nil
}
