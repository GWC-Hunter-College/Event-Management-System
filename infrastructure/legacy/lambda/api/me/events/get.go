package main

import (
	// models "cdk-infrastructure/database/models"

	"cdk-infrastructure/database/models"
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	event_schema "cdk-infrastructure/lambda/api/events/schema"
	authentication_utils "cdk-infrastructure/utils/auth"
	"cdk-infrastructure/utils/query_client"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

	// All of these parameters are optional
	startDate := event.QueryStringParameters["startDate"]
	endDate := event.QueryStringParameters["endDate"]
	limit := event.QueryStringParameters["limit"]
	page := event.QueryStringParameters["page"]

	if startDate == "" {
		startDate = "1970-01-01"
	}

	if endDate == "" {
		endDate = "2100-01-01"
	}

	// Default limit to 10 if not provided
	if limit == "" {
		limit = "10"
	}

	// TODO: Since seeded event data is less than 10 items, pagination is not fully testable yet
	// Default page to 0 if not provided
	if page == "" {
		page = "0"
	}

	// Validate date formats
	timeLayout := time.DateOnly
	if _, err := time.Parse(timeLayout, startDate); err != nil {
		return gateway_helpers.NewClientErrorGatewayResponse("Invalid startDate format. Use YYYY-MM-DD.", nil)
	}

	if _, err := time.Parse(timeLayout, endDate); err != nil {
		return gateway_helpers.NewClientErrorGatewayResponse("Invalid endDate format. Use YYYY-MM-DD.", nil)
	}

	// Validate offset
	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum <= 0 {
		return gateway_helpers.NewClientErrorGatewayResponse("Invalid limit. Must be a positive integer.", nil)
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 0 {
		return gateway_helpers.NewClientErrorGatewayResponse("Invalid page number. Must be a non-negative integer.", nil)
	}

	offset := strconv.Itoa(pageNum * limitNum)

	events := []event_schema.SQLSchema{}
	selectEventsQuery := query_client.NewQuery("students/SELECT_student_events.sql", "posted", "%", startDate, endDate, sub, limit, offset)
	err = qc.Select(&events, selectEventsQuery)

	if err != nil {
		log.Printf("Error executing query: %v", err)

		return gateway_helpers.NewServerErrorGatewayResponse("Could not fetch events: "+err.Error(), nil)
	}

	// Populate resulting events
	eventsToResponseEvent := map[int]event_schema.ResponseSchema{}

	for _, event := range events {
		_, exists := eventsToResponseEvent[event.EventID]

		var isOwner bool = event.IsOwner

		thumbnailUrl := "https://media.istockphoto.com/id/1495088043/vector/user-profile-icon-avatar-or-person-icon-profile-picture-portrait-symbol-default-portrait.jpg?s=612x612&w=0&k=20&c=dhV2p1JwmloBTOaGAtaA3AW1KSnjsdMt7-U_3EZElZ0="

		club := models.Club{
			ID: event.ClubID,
			// This is a placeholder thumbnail URL
			// TODO: Get a new presigned URL when club profile images are implemented
			ThumbnailURL: &thumbnailUrl,
		}

		if !exists {
			eventsToResponseEvent[event.EventID] = event_schema.ResponseSchema{
				Event:       event.Event,
				Description: event.Description,
				EventOwners: models.EventOwners{
					Owner:      models.Club{},
					Associates: []models.Club{},
				},
			}
		}

		var existingEvent event_schema.ResponseSchema = eventsToResponseEvent[event.EventID]

		if isOwner {
			existingEvent.Owner = club
		} else {
			existingEvent.Associates = append(existingEvent.Associates, club)
		}

		eventsToResponseEvent[event.EventID] = existingEvent
	}

	responseEvents := []event_schema.ResponseSchema{}

	for _, event := range eventsToResponseEvent {
		responseEvents = append(responseEvents, event)
	}

	response := map[string]any{
		"events": responseEvents,
	}

	return gateway_helpers.NewSuccessGatewayResponse(fmt.Sprintf("Succesfully fetched %d events of student %s", len(responseEvents), sub), response)
}

func main() { lambda.Start(handler) }
