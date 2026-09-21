package main

import (
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	authentication_utils "cdk-infrastructure/utils/auth"
	"cdk-infrastructure/utils/query_client"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	qc *query_client.QueryClient
)

func init() {
	dbName := os.Getenv("DB_NAME")
	arn := os.Getenv("DB_SECRET_ARN")
	host := os.Getenv("DB_HOST")

	client, err := query_client.NewClientFromHost(context.Background(), arn, dbName, host)
	if err != nil {
		log.Printf("Error creating query client: %v", err)
		panic(err)
	}

	qc = client
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract sub and email from JWT
	sub, email, err := authentication_utils.ExtractSubFromRequest(request.RequestContext)
	if err != nil {
		log.Printf("Error extracting sub from request: %v", err)
		return gateway_helpers.NewClientErrorGatewayResponse(
			fmt.Sprintf("Error extracting sub from request: %v", err),
			nil,
		)
	}

	// Ensure student exists (or upsert if needed)
	if err := authentication_utils.RequireStudent(ctx, qc, sub, email); err != nil {
		log.Printf("Error ensuring student exists: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			fmt.Sprintf("Error ensuring student exists: %v", err),
			nil,
		)
	}

	// Validate clubId path parameter
	clubIdStr := request.PathParameters["clubId"]
	if clubIdStr == "" {
		return gateway_helpers.NewClientErrorGatewayResponse(
			"clubId is required in path parameters",
			nil,
		)
	}

	clubId, err := strconv.Atoi(clubIdStr)
	if err != nil {
		log.Printf("Error converting clubId to integer: %v", err)
		return gateway_helpers.NewClientErrorGatewayResponse(
			"clubId must be a valid integer",
			nil,
		)
	}

	// Join club. SQL should:
	//   - ensure student and club exist
	//   - avoid duplicate membership
	//   - return 0 rows affected if join is not possible
	joinQuery := query_client.NewQuery(
		"clubs/INSERT_club_member.sql",
		clubId, // matches `JOIN clubs c ON c.id = ?`
		sub,    // matches `WHERE s.id = ?`
	)

	result, err := qc.Exec(joinQuery)
	if err != nil {
		log.Printf("Error executing club member insert query: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			"Error joining club",
			nil,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for join club: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			"Error joining club",
			nil,
		)
	}

	if rows == 0 {
		// No insert: either club does not exist, student does not exist,
		// or user is already a member.
		return gateway_helpers.NewClientErrorGatewayResponse(
			"Unable to join club. You may already be a member or the club does not exist.",
			nil,
		)
	}

	log.Printf("Student %s joined club %d", sub, clubId)

	response := map[string]any{
		"clubId": clubId,
		"joined": true,
	}

	return gateway_helpers.NewSuccessGatewayResponse(
		"Successfully joined club",
		response,
	)
}

func main() {
	lambda.Start(handler)
}
