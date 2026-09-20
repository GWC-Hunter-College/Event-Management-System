// Package client executes the extracted MySQL queries without provider-specific
// credential lookup or application policy. No connection is opened at import time.
package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/GWC-Hunter-College/Event-Management-System/database/queries"
)

// Client owns a database connection pool. Close it when its owner shuts down.
type Client struct {
	conn *sqlx.DB
}

// Open creates a lazy MySQL pool from caller-supplied configuration. It does not
// dial or ping a server. Start with mysql.NewConfig() for the driver's defaults;
// the caller supplies credentials, network/address, database name and TLS options.
func Open(config mysql.Config) (*Client, error) {
	connector, err := mysql.NewConnector(&config)
	if err != nil {
		return nil, err
	}
	return &Client{conn: sqlx.NewDb(sql.OpenDB(connector), "mysql")}, nil
}

// Ping explicitly checks connectivity using the caller's context.
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.PingContext(ctx)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Query identifies an embedded SQL file relative to database/queries/ and its
// positional arguments. SQL contents and argument ordering are preserved.
type Query struct {
	Filepath string
	Args     []any
}

func NewQuery(filepath string, args ...any) Query {
	return Query{Filepath: filepath, Args: args}
}

// Get scans the first row into dest, returning sql.ErrNoRows if none is found.
func (c *Client) Get(ctx context.Context, dest any, query Query) error {
	statement, err := queries.Load(query.Filepath)
	if err != nil {
		return err
	}
	return c.conn.GetContext(ctx, dest, statement, query.Args...)
}

// Select scans all returned rows into the slice pointed to by dest.
func (c *Client) Select(ctx context.Context, dest any, query Query) error {
	statement, err := queries.Load(query.Filepath)
	if err != nil {
		return err
	}
	return c.conn.SelectContext(ctx, dest, statement, query.Args...)
}

// Exec executes one statement and returns its affected rows and insert ID.
func (c *Client) Exec(ctx context.Context, query Query) (sql.Result, error) {
	return exec(ctx, c.conn, query)
}

func exec(ctx context.Context, conn sqlx.ExecerContext, query Query) (sql.Result, error) {
	statement, err := queries.Load(query.Filepath)
	if err != nil {
		return nil, err
	}
	return conn.ExecContext(ctx, statement, query.Args...)
}

// ExecMulti executes statements sequentially in one transaction. Results are
// returned only after a successful commit; any failure is returned to the caller.
func (c *Client) ExecMulti(ctx context.Context, queries []Query) ([]sql.Result, error) {
	var results []sql.Result
	err := c.withTx(ctx, func(tx *sqlx.Tx) error {
		for _, query := range queries {
			result, err := exec(ctx, tx, query)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ExecInsertQuery executes an initial insert and subsequent statements atomically.
// needID[i] prepends the first insert ID to queries[i+1]'s arguments. The queries
// and their argument slices are never modified. No ID is returned on failure.
func (c *Client) ExecInsertQuery(ctx context.Context, queries []Query, needID []bool) (int64, error) {
	if len(queries) == 0 {
		return 0, errors.New("no queries provided")
	}
	if len(queries) != len(needID)+1 {
		return 0, errors.New("needID must have one entry for each query after the first")
	}
	var insertID int64
	err := c.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := exec(ctx, tx, queries[0])
		if err != nil {
			return err
		}
		insertID, err = result.LastInsertId()
		if err != nil {
			return err
		}
		for i, query := range queries[1:] {
			if needID[i] {
				query.Args = append([]any{insertID}, query.Args...)
			}
			if _, err := exec(ctx, tx, query); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return insertID, nil
}

func (c *Client) withTx(ctx context.Context, run func(*sqlx.Tx) error) (err error) {
	tx, err := c.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	// Also rolls back if run panics. After Commit, database/sql returns ErrTxDone.
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback: %w", rollbackErr))
		}
	}()
	if err = run(tx); err != nil {
		return err
	}
	return tx.Commit()
}
