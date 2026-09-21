package main

import (
	"cdk-infrastructure/database/models"
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	authentication_utils "cdk-infrastructure/utils/auth"
	validation_error "cdk-infrastructure/utils/errors"
	"cdk-infrastructure/utils/query_client"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"
)

var (
	qc       *query_client.QueryClient
	validate *validator.Validate
)

func init() {
	validate = validator.New()

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

type RequestBodySchema struct {
	models.ClubDetailed `json:"club" validate:"required"`
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	sub, email, err := authentication_utils.ExtractSubFromRequest(request.RequestContext)

	if err != nil {
		log.Printf("Error extracting sub from request: %v", err)

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Error extracting sub from request: %v", err), nil)
	}

	log.Printf("Extracted sub: %s, email: %s", sub, email)

	err = authentication_utils.RequireStudent(ctx, qc, sub, email)

	if err != nil {
		log.Printf("Error requiring student: %v", err)

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Error requiring student: %v", err), nil)
	}

	// Get body and validate
	body := RequestBodySchema{}
	err = json.Unmarshal([]byte(request.Body), &body)

	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Error parsing request body: %v", err), nil)
	}

	if err = validate.Struct(body); err != nil {
		log.Printf("Validation error: %v", err)

		if errs, ok := err.(validator.ValidationErrors); ok {
			resp := validation_error.FormatValidationError(errs)

			return gateway_helpers.NewClientErrorGatewayResponse("Validation Error: Invalid request body.", resp)
		}

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Invalid request body: %v", err), nil)
	}

	insertClubQueries := []query_client.Query{
		query_client.NewQuery("clubs/INSERT_club.sql", body.Name),
		query_client.NewQuery("clubs/INSERT_club_info.sql", body.WebsiteURL, body.Description),
	}

	lastInsertId, err := qc.ExecInsertQuery(insertClubQueries, []bool{true})

	if err != nil {
		log.Printf("Error inserting club into database: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse(fmt.Sprintf("Error inserting club into database: %v", err), nil)
	}

	response := map[string]any{
		"clubId": lastInsertId,
	}

	return gateway_helpers.NewSuccessGatewayResponse("Successfully inserted club into database", response)
}

func main() {
	lambda.Start(handler)
}
