// Package queries loads the application SQL copied from the legacy query client.
package queries

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed students/ensure/*.sql me/get/*.sql me/clubs/list/*.sql me/events/list/*.sql
//go:embed clubs/list/*.sql clubs/get/*.sql clubs/create/*.sql clubs/members/create/*.sql clubs/members/leave/*.sql clubs/events/list/*.sql
//go:embed events/read/*.sql events/create/*.sql events/images/confirm/*.sql
//go:embed authorization/clubs/can_manage/*.sql authorization/events/can_manage/*.sql
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
