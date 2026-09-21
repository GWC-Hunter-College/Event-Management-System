package models_test

import (
	"context"
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GWC-Hunter-College/Event-Management-System/database/models"
	"github.com/jmoiron/sqlx"
)

// scanFixture uses a mock driver exclusively. Its query is a fixture label, not
// application SQL, and it never creates a MySQL connection.
func scanFixture(t *testing.T, dest any, columns []string, values ...driver.Value) error {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Errorf("close mock: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	mock.ExpectQuery("model scan fixture").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(values...)).
		RowsWillBeClosed()
	return sqlx.NewDb(db, "sqlmock").GetContext(context.Background(), dest, "model scan fixture")
}

func stringPointer(value string) *string { return &value }

func TestModelScanCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		values  []driver.Value
		dest    any
		want    any
	}{
		{
			name:    "student email",
			columns: []string{"id", "email"},
			values:  []driver.Value{"student-sub", "student@example.test"},
			dest:    &models.Student{},
			want:    &models.Student{ID: "student-sub", Email: "student@example.test"},
		},
		{
			name:    "club role with embedded club and nullable thumbnail",
			columns: []string{"id", "name", "thumbnail_url", "role"},
			values:  []driver.Value{int64(7), "Club", nil, "owner"},
			dest:    &models.ClubWithRole{},
			want:    &models.ClubWithRole{Club: models.Club{ID: 7, Name: "Club"}, ClubRole: "owner"},
		},
		{
			name:    "club details with nullable fields",
			columns: []string{"id", "name", "thumbnail_url", "website_url", "description"},
			values:  []driver.Value{int64(7), "Club", nil, nil, nil},
			dest:    &models.ClubDetailed{},
			want:    &models.ClubDetailed{Club: models.Club{ID: 7, Name: "Club"}},
		},
		{
			name:    "club details with populated pointers",
			columns: []string{"id", "name", "thumbnail_url", "website_url", "description"},
			values:  []driver.Value{int64(7), "Club", "images/logo", "https://example.test", "Description"},
			dest:    &models.ClubDetailed{},
			want: &models.ClubDetailed{
				Club:        models.Club{ID: 7, Name: "Club", ThumbnailURL: stringPointer("images/logo")},
				WebsiteURL:  stringPointer("https://example.test"),
				Description: stringPointer("Description"),
			},
		},
		{
			name:    "full event embedded fields and implicit description mapping",
			columns: []string{"id", "fk_author_id", "title", "description"},
			values:  []driver.Value{int64(8), "student-sub", "Event", "Description"},
			dest:    &models.FullEvent{},
			want: &models.FullEvent{
				Event:       models.Event{EventID: 8, AuthorID: "student-sub", Title: "Event"},
				Description: stringPointer("Description"),
			},
		},
		{
			name:    "event description nullable field and foreign key tag",
			columns: []string{"fk_event_id", "description"},
			values:  []driver.Value{int64(8), nil},
			dest:    &models.EventDescription{},
			want:    &models.EventDescription{EventID: 8},
		},
		{
			name:    "event image foreign keys retain string types",
			columns: []string{"fk_event_id", "fk_image_id"},
			values:  []driver.Value{int64(8), "image-id"},
			dest:    &models.EventImage{},
			want:    &models.EventImage{EventID: "8", ImageID: "image-id"},
		},
		{
			name:    "image tags and string timestamp",
			columns: []string{"id", "purpose", "object_key", "filename", "mimetype", "created_at"},
			values:  []driver.Value{"image-id", "event", "events/image", "image.png", "image/png", []byte("2025-11-04 09:00:00")},
			dest:    &models.Image{},
			want: &models.Image{
				ImageID: "image-id", Purpose: "event", ObjectKey: "events/image",
				Filename: "image.png", Type: "image/png", CreatedAt: "2025-11-04 09:00:00",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := scanFixture(t, test.dest, test.columns, test.values...); err != nil {
				t.Fatalf("scan: %v", err)
			}
			if !reflect.DeepEqual(test.dest, test.want) {
				t.Errorf("scan = %#v, want %#v", test.dest, test.want)
			}
		})
	}
}

func TestEventScanNullablePointersAndStringDates(t *testing.T) {
	for _, populated := range []bool{false, true} {
		name := "null pointers"
		var thumbnail, rsvp, deleted driver.Value
		want := models.Event{
			EventID: 8, AuthorID: "student-sub", Title: "Event", Location: "Campus",
			Status: "posted", StartDate: "2025-11-04 09:00:00", EndDate: "2025-11-04 10:00:00",
			Timezone: "America/New_York", CreatedAt: "2025-11-01 08:00:00", UpdatedAt: "2025-11-02 08:00:00",
		}
		if populated {
			name = "populated pointers"
			thumbnail, rsvp, deleted = "image-id", "https://example.test/rsvp", "2025-11-05 08:00:00"
			want.ThumbnailID = stringPointer(thumbnail.(string))
			want.RsvpLink = stringPointer(rsvp.(string))
			want.DeletedAt = stringPointer(deleted.(string))
		}
		t.Run(name, func(t *testing.T) {
			var got models.Event
			err := scanFixture(t, &got, []string{
				"id", "fk_author_id", "fk_thumbnail_id", "title", "location", "rsvp_link", "status",
				"start_date", "end_date", "timezone", "created_at", "updated_at", "deleted_at",
			}, int64(8), "student-sub", thumbnail, "Event", "Campus", rsvp, "posted",
				[]byte(want.StartDate), []byte(want.EndDate), want.Timezone,
				[]byte(want.CreatedAt), []byte(want.UpdatedAt), deleted)
			if err != nil {
				t.Fatalf("scan: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("scan = %#v, want %#v", got, want)
			}
		})
	}
}

// These failures document legacy scan behavior. Phase 1 deliberately preserves
// the nonnullable model fields even where schema columns permit NULL.
func TestExistingNonnullableStringScanFailures(t *testing.T) {
	tests := []struct {
		name   string
		column string
		dest   any
	}{
		{name: "student email", column: "email", dest: &models.Student{}},
		{name: "club name", column: "name", dest: &models.Club{}},
	}
	for _, column := range []string{
		"fk_author_id", "title", "location", "status", "start_date", "end_date", "timezone", "created_at", "updated_at",
	} {
		tests = append(tests, struct {
			name   string
			column string
			dest   any
		}{name: "event " + column, column: column, dest: &models.Event{}})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := scanFixture(t, test.dest, []string{test.column}, nil)
			if err == nil || !strings.Contains(err.Error(), "converting NULL to string is unsupported") {
				t.Fatalf("scan NULL into %s: got %v, want existing NULL-to-string error", test.column, err)
			}
		})
	}
}
