// lambdas/db-test/main.go
package main

//=============================================
// Imports
//=============================================
import (
	"context"       // carries cancellation / deadlines across calls
	"database/sql"  // std-lib DB abstraction
	"encoding/json" // (un)marshal Secrets Manager JSON payload
	"fmt"           // string formatting
	"log"           // structured Lambda logging
	"os"            // read environment variables
	"sync"          // sync.Once for cold-start cache

	"github.com/aws/aws-lambda-go/events" // API Gateway V2 types
	"github.com/aws/aws-lambda-go/lambda" // Lambda bootstrap
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	// loads AWS creds/region
	_ "github.com/go-sql-driver/mysql" // MySQL driver; blank import means “register”
)

//=============================================
// Global cache (warm-start optimisation)
//=============================================

// ❯ Why global?  A Lambda execution environment can be reused for many requests
//
//	(“warm invocation”).  Using a package-level variable lets us avoid fetching
//	the secret on every call.
//
// ❯ sync.Once guarantees loadSecret() runs exactly once per warm container.
var (
	once   sync.Once
	dbUser string
	dbPass string

	dbPort int
	dbName string

	secErr error // remember the first error so later calls return it
)

// =============================================
// loadSecret: fetches & caches username/password from Secrets Manager
// =============================================
func loadSecret(ctx context.Context, arn string) error {
	once.Do(func() { // executes only the first time
		// AWS SDK config (region/creds from env/IAM role)
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			secErr = err
			return
		}
		sm := secretsmanager.NewFromConfig(cfg)

		// Get secret value by ARN
		out, err := sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
			SecretId: &arn,
		})
		if err != nil {
			secErr = err
			return
		}

		// Parse
		var tmp struct {
			User string `json:"username"`
			Pass string `json:"password"`

			// Host string `json:"host"`
			Port int    `json:"port"`
			Name string `json:"dbname"`
		}
		if err := json.Unmarshal([]byte(*out.SecretString), &tmp); err != nil {
			secErr = err
			return
		}

		// Cache for the life of the container
		dbUser, dbPass, dbPort, dbName =
			tmp.User, tmp.Pass, tmp.Port, tmp.Name
	})
	return secErr // nil on success, first error otherwise
}

// =============================================
//
//	Response helper (serialises to JSON)
//
// =============================================
type resp struct {
	Success bool   `json:"success"`
	Result  int    `json:"result,omitempty"` // present only on OK
	Error   string `json:"error,omitempty"`  // present only on error
}

// =============================================
// Lambda handler – runs every invocation
// =============================================
func handler(ctx context.Context, evt events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	// arn of secret
	arn := os.Getenv("DB_SECRET_ARN") // secret reference

	host := os.Getenv("DB_HOST") // secret reference

	// fet credentionals from secret
	if err := loadSecret(ctx, arn); err != nil {
		return fail(err) // early return on error
	}

	// tls stuff needed

	// Build MySQL DSN and connect
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=true", dbUser, dbPass, host, dbPort, dbName)
	db, err := sql.Open("mysql", dsn) // creates connection pool
	if err != nil {
		return fail(err)
	}
	defer db.Close()

	// Simple test query
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1 + 1 AS result").Scan(&result); err != nil {
		return fail(err)
	}

	// Marshal success JSON
	ok, _ := json.Marshal(resp{Success: true, Result: result})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(ok),
	}, nil
}

// =============================================
// error wrapper
// =============================================
func fail(e error) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("handler error:", e) // appears in CloudWatch Logs
	body, _ := json.Marshal(resp{Success: false, Error: e.Error()})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 500,
		Body:       string(body),
	}, nil
}

// =============================================
// Bootstrap
// =============================================
func main() {
	lambda.Start(handler)
}
