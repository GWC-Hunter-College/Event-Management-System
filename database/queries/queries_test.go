package queries

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadMigratedQueries(t *testing.T) {
	if len(migratedQueries) != 21 {
		t.Fatalf("inventory contains %d queries, want 21", len(migratedQueries))
	}
	wantPaths := make(map[string]bool, len(migratedQueries))
	for _, query := range migratedQueries {
		wantPaths[query.path] = true
		t.Run(query.path, func(t *testing.T) {
			got, err := Load(query.path)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != query.size {
				t.Errorf("loaded %d bytes, want original %d bytes", len(got), query.size)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != query.sha256 {
				t.Errorf("SQL differs from original bytes: SHA256 %s, want %s", hash, query.sha256)
			}
			if strings.ContainsRune(got, '\x00') {
				t.Error("loaded SQL contains null bytes")
			}
		})
	}

	count := 0
	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		count++
		if !wantPaths[path] {
			t.Errorf("unexpected embedded file %q", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != len(migratedQueries) {
		t.Errorf("embedded %d files, want %d", count, len(migratedQueries))
	}
}

func TestLoadCompleteFiles(t *testing.T) {
	for _, size := range []int{0, 1, 4095, 4096, 4097, 16384} {
		t.Run(fmt.Sprintf("%d_bytes", size), func(t *testing.T) {
			want := strings.Repeat("x", size)
			source := fstest.MapFS{"query.sql": &fstest.MapFile{Data: []byte(want)}}
			got, err := load(source, "query.sql")
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("load changed content: got %d bytes, want %d", len(got), len(want))
			}
		})
	}
}

func TestLoadUnavailableFile(t *testing.T) {
	for _, path := range []string{
		"events/images/list/SELECT_event_images.sql",
		"../README.md",
		"/students/ensure/UPSERT_student.sql",
	} {
		t.Run(path, func(t *testing.T) {
			got, err := Load(path)
			if err == nil {
				t.Fatal("load succeeded for an unavailable query")
			}
			if got != "" {
				t.Errorf("failed load returned content %q", got)
			}
			if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, fs.ErrInvalid) {
				t.Errorf("load did not preserve filesystem error: %v", err)
			}
		})
	}
}
