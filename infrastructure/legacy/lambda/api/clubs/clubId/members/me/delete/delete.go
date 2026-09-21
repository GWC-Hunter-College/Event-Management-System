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

	// Ensure student exists
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

	// Leave club. SQL should:
	//   - delete membership only if not owner
	//   - return 0 rows affected if user is not a member or is owner
	leaveQuery := query_client.NewQuery(
		"clubs/DELETE_club_member.sql",
		sub,    // fk_student_id = ?
		clubId, // fk_club_id = ?
	)

	result, err := qc.Exec(leaveQuery)
	if err != nil {
		log.Printf("Error executing club member delete query: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			"Error leaving club",
			nil,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for leave club: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			"Error leaving club",
			nil,
		)
	}

	if rows == 0 {
		// Either not a member, or they are the owner and the SQL protected the row.
		return gateway_helpers.NewClientErrorGatewayResponse(
			"Unable to leave club. You may not be a member or you are the club owner.",
			nil,
		)
	}

	log.Printf("Student %s left club %d", sub, clubId)

	response := map[string]any{
		"clubId": clubId,
		"left":   true,
	}

	return gateway_helpers.NewSuccessGatewayResponse(
		"Successfully left club",
		response,
	)
}

func main() {
	lambda.Start(handler)
}
