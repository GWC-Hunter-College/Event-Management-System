// query_client.go
// A client to streamline connecting to  and querying the deployed RDS database.

package query_client

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type QueryClient struct {
	Conn *sqlx.DB

	user     string
	password string
	address  string
}

// The below go:embed directive will embed all .sql files located in the queries directory
// into the Go binary at compile time, allowing for easy access to SQL query files WITHOUT needing to
// create a separate Dockerfile for each Lambda.

// This may seem like a lot of extra bytes added to the binary, but all files are zipped in the binary
// and only read when it's needed.

// The marginal size increase is negligible compared to the scalability of not needing to manage
// separate Dockerfiles for each Lambda function that needs to run SQL queries.
// For more information, please see: https://pkg.go.dev/embed
// - Anthony 09/08/2025

//go:embed queries/**/*.sql
var sqlFiles embed.FS

var queriesDir, err = fs.Sub(sqlFiles, "queries")

// Factory method to create a new QueryClient instance.
// It requires a context and the ARN of the AWS Secrets Manager secret.
//
// Please remember to close the client connection using `defer queryClient.Close()`
//
// Please also note that if using the client in AWS Lambda, the lambda's IAM role must have
// permissions to access the secret in AWS Secrets Manager.
func NewClient(ctx context.Context, dbSecretArn string, dbName string) (*QueryClient, error) {
	dsn := mysql.NewConfig()

	creds, err := loadSecrets(ctx, dbSecretArn)
	if err != nil {
		return nil, err
	}

	dsn.User = creds.user
	dsn.Passwd = creds.password
	dsn.Addr = creds.address
	dsn.Net = "tcp"
	// dsn.TLSConfig = "true"

	conn, err := sqlx.Connect("mysql", dsn.FormatDSN())

	if err != nil {
		return nil, err
	}

	return &QueryClient{
		Conn:     conn,
		user:     creds.user,
		password: creds.password,
		address:  creds.address,
	}, nil
}

// Factory method to create a new QueryClient instance with a custom host address.
// It requires a context and the ARN of the AWS Secrets Manager secret, along with the custom host address.
//
// Please remember to close the client connection using `defer queryClient.Close()`
//
// Please also note that if using the client in AWS Lambda, the lambda's IAM role must have
// permissions to access the secret in AWS Secrets Manager.
func NewClientFromHost(ctx context.Context, dbSecretArn string, dbName string, host string) (*QueryClient, error) {
	dsn := mysql.NewConfig()

	creds, err := loadSecrets(ctx, dbSecretArn)
	if err != nil {
		return nil, err
	}

	dsn.User = creds.user
	dsn.Passwd = creds.password
	dsn.Addr = fmt.Sprintf("%s:%s", host, creds.port)
	dsn.Net = "tcp"
	dsn.DBName = dbName
	dsn.TLSConfig = "true"

	conn, err := sqlx.Connect("mysql", dsn.FormatDSN())

	if err != nil {
		return nil, err
	}

	return &QueryClient{
		Conn:     conn,
		user:     creds.user,
		password: creds.password,
		address:  creds.address,
	}, nil
}

func (qc *QueryClient) ChangeDatabase(dbName string) error {
	_, err := qc.Conn.Exec("USE " + dbName)
	if err != nil {
		return err
	}

	return nil
}

// Execute a query that can only return one row from a file located at `filepath` with optional args for placeholders.
// This method automatically scans the returned rows into `dest`.
//
// This method automatically scans the returned rows into `dest`. It should be the address to an interface instance.
//
// Example:
//
//	var event = models.Event{}
//	selectEventQuery := query_client.NewQuery("events/SELECT_event.sql", arg1, arg2)
//	err := qc.Select(&event, selectEventQuery)
func (qc *QueryClient) Get(dest interface{}, query Query) error {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return err
	}

	err = qc.Conn.Get(dest, queryString, query.Args...)
	if err != nil {
		log.Printf("Error executing query from file %s: %v", query.Filepath, err)
		return err
	}

	log.Printf("Successfully executed query from file: %s.", query.Filepath)

	return nil
}

// Execute a query that can return multiple rows from a file located at `filepath` with optional args for placeholders.
//
// This method automatically scans the returned rows into `dest`. It should be the address of a slice of interface instances.
//
// Example:
//
//	var events = []models.Event{}
//	selectEventsQuery := query_client.NewQuery("events/SELECT_events.sql", arg1, arg2)
//	err := qc.Select(&events, selectEventsQuery)
func (qc *QueryClient) Select(dest interface{}, query Query) error {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return err
	}

	err = qc.Conn.Select(dest, queryString, query.Args...)
	if err != nil {
		log.Printf("Error executing query from file %s: %v", query.Filepath, err)
		return err
	}

	log.Printf("Successfully executed query from file: %s", query.Filepath)

	return nil
}

// Execute a query from a file located at `filepath` with optional args for placeholders.
//
// This is not expected to return any rows, and is typically used for INSERT, UPDATE, DELETE, or DDL statements.
func (qc *QueryClient) Exec(query Query) (sql.Result, error) {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return nil, err
	}

	result, err := qc.Conn.Exec(queryString, query.Args...)
	if err != nil {
		log.Printf("Error executing query from file %s: %v", query.Filepath, err)
		return result, err
	}

	log.Printf("Successfully executed query from file: %s", query.Filepath)

	return result, nil
}

// Execute a query from a file located at `filepath` that is expected to return a single row.
//
// This is the manual version of QueryClient.Get() that returns a sqlx.Row for custom scanning.
func (qc *QueryClient) QueryRow(query Query) *sqlx.Row {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return nil
	}

	row := qc.Conn.QueryRowx(queryString, query.Args...)

	log.Printf("Successfully executed query from file: %s.", query.Filepath)

	return row
}

// Execute a query from a file located at `filepath` that is expected to return multiple rows.
//
// This is the manual version of QueryClient.Select() that returns sqlx.Rows for custom scanning.
func (qc *QueryClient) Query(query Query) (*sqlx.Rows, error) {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return nil, err
	}

	rows, err := qc.Conn.Queryx(queryString, query.Args...)
	if err != nil {
		log.Printf("Error executing query from file %s: %v", query.Filepath, err)
		return nil, err
	}

	log.Printf("Successfully executed query from file: %s", query.Filepath)

	return rows, nil
}

// Execute multiple queries in sequence, each from their own file with optional args for placeholders in a transaction.
// If any query fails, the transaction is aborted and rolled back.
func (qc *QueryClient) QueryMulti(queries []Query) (rows []*sqlx.Rows, err error) {
	tx, err := qc.Conn.Beginx()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return nil, err
	}

	for _, query := range queries {
		queryString, err := loadSQLFromFile(query.Filepath)
		if err != nil {
			log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
			tx.Rollback()
			return nil, err
		}

		resultingRows, err := tx.Queryx(queryString, query.Args...)

		log.Printf("Executed query from file: %s", query.Filepath)

		if err != nil {
			log.Printf("Error executing query from file %s: %v", query.Filepath, err)

			tx.Rollback()

			return rows, err
		}

		rows = append(rows, resultingRows)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	return rows, nil
}

// Execute multiple queries in sequence, each from their own file with optional args for placeholders in a transaction.
// If any query fails, the transaction is aborted and rolled back.
//
// This method does NOT return any rows, meaning that this method should only be used
// to execute statements that do not return rows (e.g. CREATE, INSERT, UPDATE, DELETE).
func (qc *QueryClient) ExecMulti(queries []Query) (results []sql.Result, err error) {
	tx, err := qc.Conn.Beginx()

	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return results, err
	}

	for _, query := range queries {
		queryString, err := loadSQLFromFile(query.Filepath)
		if err != nil {
			log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
			tx.Rollback()
			return nil, err
		}

		result, err := tx.Exec(queryString, query.Args...)

		if err != nil {
			log.Printf("Error executing query from file %s: %v", query.Filepath, err)

			tx.Rollback()

			return results, err
		}

		log.Printf("Executed query from file: %s", query.Filepath)

		results = append(results, result)
	}

	err = tx.Commit()

	return results, nil
}

// Execute multiple queries in sequence, each from their own file with optional args for placeholders in a transaction.
// If any query fails, the transaction is aborted and rolled back.
//
// This variation of Exec uses the result of the first query to return the last inserted ID.
// This is useful for cases where you need to insert a record and then use its ID for subsequent queries.
//
// Please note that the first query must be an INSERT statement that returns a valid last inserted ID.
//
// The needId slice indicates which queries require the last inserted ID as the first argument.
// The first boolean in needId corresponds to the second query, the second boolean corresponds to the third query, and so on.
// Therefore, this method expects len(queries) == len(needId) + 1.
func (qc *QueryClient) ExecInsertQuery(queries []Query, needId []bool) (lastInsertId int64, err error) {
	if len(queries) == 0 {
		return 0, errors.New("no queries provided")
	}

	if len(queries) != len(needId)+1 {
		return 0, errors.New("length of queries and needId must be the same")
	}

	tx, err := qc.Conn.Beginx()

	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return 0, err
	}

	initialQueryString, err := loadSQLFromFile(queries[0].Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", queries[0].Filepath, err)
		tx.Rollback()
		return 0, err
	}

	insertMainResult, err := tx.Exec(initialQueryString, queries[0].Args...)
	if err != nil {
		log.Printf("Error executing query from file %s: %v", queries[0].Filepath, err)
		tx.Rollback()
		return 0, err
	}

	lastInsertId, err = insertMainResult.LastInsertId()
	if err != nil {
		log.Printf("Error getting last inserted ID: %v", err)
		tx.Rollback()
		return 0, err
	}

	log.Printf("Executed query from file: %s", queries[0].Filepath)

	for i, query := range queries[1:] {
		queryString, err := loadSQLFromFile(query.Filepath)
		if err != nil {
			log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
			tx.Rollback()
			return 0, err
		}

		var args []any
		if needId[i] {
			args = append([]any{lastInsertId}, query.Args...)
		} else {
			args = query.Args
		}

		_, err = tx.Exec(queryString, args...)

		if err != nil {
			log.Printf("Error executing query from file %s: %v", query.Filepath, err)

			tx.Rollback()

			return 0, err
		}

		log.Printf("Executed query from file: %s", query.Filepath)
	}

	err = tx.Commit()

	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return 0, err
	}

	return lastInsertId, nil
}

// DO NOT USE THIS METHOD. IT DOES NOT WORK PROPERLY.
//
// Execute a SQL file that contains multiple statements (e.g. for migrations or bulk table creation).
// Each statement is executed in sequence and atomically in a transaction.
//
// Please note that this method does NOT return any rows, meaning that this method should only be used
// to execute statements that do not return rows (e.g. CREATE, INSERT, UPDATE, DELETE).
//
// The verbose flag can be set to true to log each statement execution. Please note that this will log
// everything, including sensitive information if the statements contain such information.
/*
func (qc *QueryClient) ExecuteFileBulk(query Query) ([]sql.Result, error) {
	queryString, err := loadSQLFromFile(query.Filepath)
	if err != nil {
		log.Printf("Error loading SQL from file %s: %v", query.Filepath, err)
		return nil, err
	}

	var results = []sql.Result{}

	tx, err := qc.Conn.Beginx()

	statements := strings.Split(queryString, ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		result, err := tx.Exec(statement)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			tx.Rollback()
			return nil, err
		}

		results = append(results, result)

		log.Printf("Executed statement: %s", statement)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	return results, nil
}
*/

func (qc *QueryClient) Close() error {
	return qc.Conn.Close()
}
