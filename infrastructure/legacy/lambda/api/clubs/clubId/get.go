package main

import (
	models "cdk-infrastructure/database/models"
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	"cdk-infrastructure/utils/query_client"
	"context"
	"database/sql"
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
	clubId := request.PathParameters["clubId"]

	club := models.ClubDetailed{}
	selectClubsQuery := query_client.NewQuery("clubs/SELECT_club.sql", clubId)
	err := qc.Get(&club, selectClubsQuery)

	if err != nil {
		if err == sql.ErrNoRows {
			// not found → return 404
			return gateway_helpers.NewClientErrorGatewayResponse(
				fmt.Sprintf("Club with id %s not found", clubId),
				nil,
			)
		}

		// DB error
		log.Printf("Error executing query: %v", err)
		return gateway_helpers.NewServerErrorGatewayResponse(
			"Could not fetch club: "+err.Error(),
			nil,
		)
	}

	response := map[string]any{
		"club": club,
	}

	return gateway_helpers.NewSuccessGatewayResponse(fmt.Sprintf("Succesfully fetched club %s", clubId), response)
}

func main() {
	lambda.Start(handler)
}
