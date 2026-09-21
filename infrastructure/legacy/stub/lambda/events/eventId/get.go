package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"cdk-infrastructure/stub/shared"
)

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Stubbed data (e.g., for kyle@example.com)
	eventID := req.PathParameters["eventId"]

	resp := shared.EventDetailed{
		ID:          1,
		Title:       "HunterHacks",
		Location:    "Room 101, Hunter College",
		RSVPLink:    "https://www.hunterhacks.com/",
		Status:      shared.StatusPosted,
		StartISO:    "2025-09-15 17:00:00",
		EndISO:      "2025-09-15 19:00:00",
		Timezone:    "America/New_York",
		CreateISO:   "2025-09-05 19:00:00",
		UpdateISO:   "2025-09-05 19:00:00",
		Description: fmt.Sprintf("Event ID you inputed is: %s", eventID),
	}

	body, err := json.Marshal(resp)
	if err != nil {
		errBody, _ := json.Marshal(shared.ErrorResponse{Error: "failed_to_marshal"})
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(errBody),
		}, err
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handle)
}
