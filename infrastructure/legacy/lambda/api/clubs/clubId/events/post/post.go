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
	"strconv"

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
	Event       models.Event `json:"event"`
	Description string       `json:"description" validate:"required"`
	Associates  []int        `json:"associates" validate:"required"`
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	sub := ""
	email := ""

	sub, email, err := authentication_utils.ExtractSubFromRequest(request.RequestContext)

	if err != nil {
		log.Printf("Error extracting sub from request: %v", err)

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Error extracting sub from request: %v", err), nil)
	}

	if err := authentication_utils.RequireStudent(ctx, qc, sub, email); err != nil {
		log.Printf("Error ensuring student exists: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse(fmt.Sprintf("Error ensuring student exists: %v", err), nil)
	}

	// Verify clubId is a valid integer
	clubIdStr := request.PathParameters["clubId"]

	if clubIdStr == "" {
		return gateway_helpers.NewClientErrorGatewayResponse("clubId is required in path parameters", nil)
	}

	clubId, err := strconv.Atoi(clubIdStr)
	if err != nil {
		log.Printf("Error converting clubId to integer: %v", err)

		return gateway_helpers.NewClientErrorGatewayResponse("clubId must be a valid integer", nil)
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

			return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Validation Error: Invalid request body."), resp)
		}

		return gateway_helpers.NewClientErrorGatewayResponse(fmt.Sprintf("Invalid request body: %v", err), nil)
	}

	// Start inserts
	insertQuery := query_client.NewQuery("events/INSERT_event.sql", []any{
		sub,
		body.Event.Title,
		body.Event.StartDate,
		body.Event.EndDate,
		body.Event.Location,
		body.Event.Timezone,
		body.Event.RsvpLink,
	}...)

	insertResult, err := qc.Exec(insertQuery)

	if err != nil {
		log.Printf("Error executing event sub queries: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse(fmt.Sprintf("Error executing event queries: %v", err), nil)
	}

	eventId, err := insertResult.LastInsertId()

	if err != nil {
		log.Printf("Error getting last insert id from insert result: %v", insertResult)

		return gateway_helpers.NewServerErrorGatewayResponse(fmt.Sprintf("Error getting last insert id from insert result: %v", err), nil)
	}

	linkQueries := []query_client.Query{
		query_client.NewQuery("events/INSERT_event_club_link.sql", []any{
			eventId,
			clubId,
			true,
		}...),
		query_client.NewQuery("events/INSERT_event_description.sql", []any{
			eventId,
			body.Description,
		}...),
	}

	for _, associateId := range body.Associates {
		if associateId == clubId {
			continue
		}

		linkQueries = append(linkQueries, query_client.NewQuery("events/INSERT_event_club_link.sql", []any{
			eventId,
			associateId,
			false,
		}...))
	}

	_, err = qc.ExecMulti(linkQueries)

	if err != nil {
		log.Printf("Error executing event-club link queries: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse(fmt.Sprintf("Error executing event-club link queries: %v", err), nil)
	}

	log.Printf("Successfully inserted event with ID %d for club %d by user %s", eventId, clubId, sub)

	response := map[string]any{
		"eventId": eventId,
	}

	return gateway_helpers.NewSuccessGatewayResponse("Successfully inserted event into database", response)
}

func main() {
	lambda.Start(handler)
}
