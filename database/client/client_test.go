package client

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"net"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/GWC-Hunter-College/Event-Management-System/database/models"
	"github.com/GWC-Hunter-College/Event-Management-System/database/queries"
)

const (
	studentPath = "me/get/SELECT_student_by_sub.sql"
	clubsPath   = "clubs/list/SELECT_clubs.sql"
	insertPath  = "clubs/create/INSERT_club.sql"
	infoPath    = "clubs/create/INSERT_club_info.sql"
)

func mockClient(t *testing.T) (*Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{conn: sqlx.NewDb(db, "mysql")}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		mock.ExpectClose()
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	return c, mock
}

func statement(t *testing.T, path string) string {
	t.Helper()
	query, err := queries.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return query
}

func TestOpenIsLazyAndPingUsesCallerContext(t *testing.T) {
	cfg := mysql.NewConfig()
	cfg.User, cfg.Passwd, cfg.DBName = "test-user", "test-password", "test-schema"
	cfg.Net, cfg.Addr = "tcp", "unused.invalid:3306"
	dialErr := errors.New("test dial blocked")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	// A stub dialer guarantees that even Ping cannot use the network.
	cfg.DialFunc = func(got context.Context, network, addr string) (net.Conn, error) {
		calls++
		if got != ctx || network != cfg.Net || addr != cfg.Addr {
			t.Errorf("dial did not receive caller context/address: %v %q %q", got, network, addr)
		}
		return nil, dialErr
	}
	c, err := Open(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if calls != 0 {
		t.Fatal("Open dialed a server")
	}
	if err := c.Ping(ctx); !errors.Is(err, dialErr) {
		t.Fatalf("Ping error = %v, want stub error", err)
	}
	if calls != 1 {
		t.Fatalf("dial calls = %d, want 1", calls)
	}
}

func TestGetSelectAndExec(t *testing.T) {
	c, mock := mockClient(t)
	ctx := context.Background()
	mock.ExpectQuery(statement(t, studentPath)).WithArgs("student-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow("student-1", "student@example.test")).RowsWillBeClosed()
	var student models.Student
	if err := c.Get(ctx, &student, NewQuery(studentPath, "student-1")); err != nil {
		t.Fatal(err)
	}
	if student.ID != "student-1" || student.Email != "student@example.test" {
		t.Fatalf("unexpected student: %#v", student)
	}
	mock.ExpectQuery(statement(t, studentPath)).WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"})).RowsWillBeClosed()
	if err := c.Get(ctx, &student, NewQuery(studentPath, "missing")); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Get missing row = %v, want sql.ErrNoRows", err)
	}
	mock.ExpectQuery(statement(t, clubsPath)).WithArgs(false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "thumbnail_url"}).
			AddRow(1, "GWC", nil).AddRow(2, "Other club", "images/logo")).RowsWillBeClosed()
	var clubs []models.Club
	if err := c.Select(ctx, &clubs, NewQuery(clubsPath, false)); err != nil {
		t.Fatal(err)
	}
	if len(clubs) != 2 || clubs[0].ID != 1 || clubs[0].ThumbnailURL != nil || clubs[1].ThumbnailURL == nil || *clubs[1].ThumbnailURL != "images/logo" {
		t.Fatalf("unexpected clubs: %#v", clubs)
	}
	mock.ExpectExec(statement(t, insertPath)).WithArgs("GWC").WillReturnResult(sqlmock.NewResult(42, 1))
	result, err := c.Exec(ctx, NewQuery(insertPath, "GWC"))
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil || id != 42 {
		t.Fatalf("insert ID = %d, %v", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		t.Fatalf("affected rows = %d, %v", affected, err)
	}
}

func TestReadErrors(t *testing.T) {
	t.Run("query error", func(t *testing.T) {
		c, mock := mockClient(t)
		want := errors.New("query failed")
		mock.ExpectQuery(statement(t, studentPath)).WithArgs("student-1").WillReturnError(want)
		var student models.Student
		if err := c.Get(context.Background(), &student, NewQuery(studentPath, "student-1")); !errors.Is(err, want) {
			t.Fatalf("Get error = %v", err)
		}
	})
	t.Run("iteration error", func(t *testing.T) {
		c, mock := mockClient(t)
		want := errors.New("rows failed")
		mock.ExpectQuery(statement(t, clubsPath)).WithArgs(false).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "GWC").RowError(0, want)).RowsWillBeClosed()
		var clubs []models.Club
		if err := c.Select(context.Background(), &clubs, NewQuery(clubsPath, false)); !errors.Is(err, want) {
			t.Fatalf("Select error = %v", err)
		}
	})
}

func TestOperationsRejectMissingSQLAndCanceledContext(t *testing.T) {
	operations := map[string]func(*Client, context.Context, string) error{
		"Get": func(c *Client, ctx context.Context, path string) error {
			var student models.Student
			return c.Get(ctx, &student, NewQuery(path))
		},
		"Select": func(c *Client, ctx context.Context, path string) error {
			var students []models.Student
			return c.Select(ctx, &students, NewQuery(path))
		},
		"Exec": func(c *Client, ctx context.Context, path string) error {
			_, err := c.Exec(ctx, NewQuery(path))
			return err
		},
	}
	for name, run := range operations {
		t.Run(name, func(t *testing.T) {
			c, _ := mockClient(t)
			if err := run(c, context.Background(), "missing.sql"); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("missing SQL error = %v", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := run(c, ctx, studentPath); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled context error = %v", err)
			}
		})
	}
}

func TestExecMultiCommit(t *testing.T) {
	c, mock := mockClient(t)
	mock.ExpectBegin()
	mock.ExpectExec(statement(t, insertPath)).WithArgs("one").WillReturnResult(sqlmock.NewResult(41, 1))
	mock.ExpectExec(statement(t, insertPath)).WithArgs("two").WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()
	results, err := c.ExecMulti(context.Background(), []Query{NewQuery(insertPath, "one"), NewQuery(insertPath, "two")})
	if err != nil || len(results) != 2 {
		t.Fatalf("ExecMulti = %v, %v", results, err)
	}
	for i, result := range results {
		if id, err := result.LastInsertId(); err != nil || id != int64(41+i) {
			t.Errorf("result %d = %d, %v", i, id, err)
		}
	}
}

func TestExecInsertQueryCommitAndArguments(t *testing.T) {
	c, mock := mockClient(t)
	args := []any{"https://example.test", "Club description"}
	batch := []Query{NewQuery(insertPath, "GWC"), NewQuery(infoPath, args...), NewQuery(insertPath, "Other")}
	mock.ExpectBegin()
	mock.ExpectExec(statement(t, insertPath)).WithArgs("GWC").WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(statement(t, infoPath)).WithArgs(int64(42), args[0], args[1]).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(statement(t, insertPath)).WithArgs("Other").WillReturnResult(sqlmock.NewResult(43, 1))
	mock.ExpectCommit()
	id, err := c.ExecInsertQuery(context.Background(), batch, []bool{true, false})
	if err != nil || id != 42 {
		t.Fatalf("ExecInsertQuery = %d, %v", id, err)
	}
	if !reflect.DeepEqual(batch[1].Args, []any{"https://example.test", "Club description"}) || !reflect.DeepEqual(batch[2].Args, []any{"Other"}) {
		t.Fatalf("caller arguments changed: %#v", batch)
	}
}

func TestTransactionFailures(t *testing.T) {
	failure := errors.New("transaction failed")
	rollbackFailure := errors.New("rollback failed")
	for _, insert := range []bool{false, true} {
		name := "ExecMulti"
		if insert {
			name = "ExecInsertQuery"
		}
		t.Run(name, func(t *testing.T) {
			for _, stage := range []string{"begin", "first load", "first exec", "second load", "second exec", "rollback", "commit", "canceled", "insert ID"} {
				if stage == "insert ID" && !insert {
					continue
				}
				t.Run(stage, func(t *testing.T) {
					c, mock := mockClient(t)
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					batch := []Query{NewQuery(insertPath, "one"), NewQuery(insertPath, "two")}
					want := failure
					switch stage {
					case "canceled":
						cancel()
						want = context.Canceled
					case "begin":
						mock.ExpectBegin().WillReturnError(failure)
					default:
						mock.ExpectBegin()
						if stage == "first load" {
							batch[0].Filepath = "missing.sql"
							want = fs.ErrNotExist
						} else {
							first := mock.ExpectExec(statement(t, insertPath)).WithArgs("one")
							if stage == "first exec" {
								first.WillReturnError(failure)
							} else if stage == "insert ID" {
								first.WillReturnResult(sqlmock.NewErrorResult(failure))
							} else {
								first.WillReturnResult(sqlmock.NewResult(42, 1))
								if stage == "second load" {
									batch[1].Filepath = "missing.sql"
									want = fs.ErrNotExist
								} else {
									second := mock.ExpectExec(statement(t, insertPath)).WithArgs("two")
									if stage == "commit" {
										second.WillReturnResult(sqlmock.NewResult(43, 1))
									} else {
										second.WillReturnError(failure)
									}
								}
							}
						}
						if stage == "commit" {
							mock.ExpectCommit().WillReturnError(failure)
						} else if stage == "rollback" {
							mock.ExpectRollback().WillReturnError(rollbackFailure)
						} else {
							mock.ExpectRollback()
						}
					}
					var err error
					if insert {
						var id int64
						id, err = c.ExecInsertQuery(ctx, batch, []bool{false})
						if id != 0 {
							t.Errorf("failed transaction returned ID %d", id)
						}
					} else {
						var results []sql.Result
						results, err = c.ExecMulti(ctx, batch)
						if results != nil {
							t.Error("failed transaction returned results")
						}
					}
					if !errors.Is(err, want) {
						t.Errorf("error = %v, want %v", err, want)
					}
					if stage == "rollback" && !errors.Is(err, rollbackFailure) {
						t.Errorf("rollback error was lost: %v", err)
					}
				})
			}
		})
	}
}

func TestExecInsertQueryRejectsInvalidBatch(t *testing.T) {
	c, _ := mockClient(t)
	for _, batch := range [][]Query{nil, {NewQuery(insertPath), NewQuery(infoPath)}} {
		if id, err := c.ExecInsertQuery(context.Background(), batch, nil); err == nil || id != 0 {
			t.Errorf("invalid batch returned %d, %v", id, err)
		}
	}
}
