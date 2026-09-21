package main

import (
	models "cdk-infrastructure/database/models"
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	"cdk-infrastructure/utils/query_client"
	"context"
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

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// All of these parameters are optional
	verified := request.QueryStringParameters["verified"]
	v := false

	if verified == "true" || verified == "TRUE" {
		v = true
	}

	clubs := []models.Club{}
	selectClubsQuery := query_client.NewQuery("clubs/SELECT_clubs.sql", v)
	err := qc.Select(&clubs, selectClubsQuery)

	if err != nil {
		log.Printf("Error executing query: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse("Could not fetch clubs: "+err.Error(), nil)
	}

	response := map[string]any{
		"clubs": clubs,
	}

	return gateway_helpers.NewSuccessGatewayResponse(fmt.Sprintf("Succesfully fetched %d clubs", len(clubs)), response)
}

func main() {
	lambda.Start(handler)
}
