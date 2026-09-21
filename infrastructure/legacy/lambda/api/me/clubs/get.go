package main

import (
	// models "cdk-infrastructure/database/models"
	"cdk-infrastructure/database/models"
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	authentication_utils "cdk-infrastructure/utils/auth"
	"cdk-infrastructure/utils/query_client"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

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

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	claims := map[string]string{}
	if event.RequestContext.Authorizer != nil &&
		event.RequestContext.Authorizer.JWT != nil &&
		event.RequestContext.Authorizer.JWT.Claims != nil {
		claims = event.RequestContext.Authorizer.JWT.Claims
	}

	sub := claims["sub"]
	email := claims["email"]

	// Enforce/ensure student row exists (or create it); fail the request if this fails.
	if err := authentication_utils.RequireStudent(ctx, qc, sub, email); err != nil {
		if errors.Is(err, authentication_utils.ErrNoSub) {
			return gateway_helpers.NewClientErrorGatewayResponse(
				"missing sub in JWT claims",
				map[string]any{"code": "ERR_NO_SUB"},
			)
		}
		// Any DB/other error -> 500
		return gateway_helpers.NewServerErrorGatewayResponse(
			"failed to ensure student",
			map[string]any{
				"code":   "ERR_REQUIRE_STUDENT",
				"detail": err.Error(),
			},
		)
	}

	// Success -> 200 with sub (and email if you want)
	myClubs := []models.ClubWithRole{}
	selectStudentClubsQuery := query_client.NewQuery("students/SELECT_student_clubs.sql", sub)
	err := qc.Select(&myClubs, selectStudentClubsQuery)

	if err != nil {
		// DB error
		// Any DB/other error -> 500
		return gateway_helpers.NewServerErrorGatewayResponse(
			"failed to query student",
			map[string]any{
				"code":   "ERR_DB_FAILURE",
				"detail": err.Error(),
			},
		)
	}

	response := map[string]any{
		"clubs": myClubs,
	}

	return gateway_helpers.NewSuccessGatewayResponse(
		fmt.Sprintf("Successfully fetched student %s", sub),
		response,
	)
}

func main() { lambda.Start(handler) }
