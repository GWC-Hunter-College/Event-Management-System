package queries

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
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

	for _, path := range embeddedPaths(t) {
		delete(wantPaths, path)
	}
	for path := range wantPaths {
		t.Errorf("pinned query %q isn't embedded", path)
	}
}

// TestNewQueriesDocumentThemselves checks that every query written in this module,
// rather than copied from legacy, starts with its header comment: the file name,
// then what it does, which endpoints use it, its parameters in order, and what it returns.
func TestNewQueriesDocumentThemselves(t *testing.T) {
	pinned := make(map[string]bool, len(migratedQueries))
	for _, query := range migratedQueries {
		if !query.fixed {
			pinned[query.path] = true
		}
	}
	for _, path := range embeddedPaths(t) {
		if pinned[path] {
			continue
		}
		t.Run(path, func(t *testing.T) {
			if !strings.HasSuffix(path, ".sql") {
				t.Fatal("embedded file isn't SQL")
			}
			got, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			name := path[strings.LastIndex(path, "/")+1:]
			if !strings.HasPrefix(got, "-- "+name+"\n") {
				t.Errorf("doesn't start with -- %s", name)
			}
			header := got[:strings.Index(got+"\n\n", "\n\n")]
			for _, label := range []string{"-- Does:", "-- Used by:", "-- Params", "-- Returns:"} {
				if !strings.Contains(header, label) {
					t.Errorf("header has no %q line", label)
				}
			}
			prefix := name[:strings.Index(name, "_")]
			switch prefix {
			case "SELECT", "INSERT", "UPDATE", "DELETE", "UPSERT", "EXISTS", "IS":
			default:
				t.Errorf("name prefix %q isn't one of the naming conventions", prefix)
			}
		})
	}
}

// TestEverySQLFileIsEmbedded catches a query directory missing from the go:embed
// list, which would make Load fail for its files at run time.
func TestEverySQLFileIsEmbedded(t *testing.T) {
	embedded := make(map[string]bool)
	for _, path := range embeddedPaths(t) {
		embedded[path] = true
	}
	err := fs.WalkDir(os.DirFS("."), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".sql") && !embedded[path] {
			t.Errorf("%s isn't in the go:embed list", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func embeddedPaths(t *testing.T) []string {
	t.Helper()
	var paths []string
	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return paths
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
		"events/images/list/SELECT_missing.sql",
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
