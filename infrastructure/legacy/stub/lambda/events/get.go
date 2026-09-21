package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"cdk-infrastructure/stub/shared"
)

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Stubbed data (e.g., for kyle@example.com)
	resp := shared.Events{
		Events: []shared.Event{
			{
				ID:       1,
				Title:    "First Kickoff Meeting",
				Location: "Room 101, Hunter College",
				RSVPLink: "https://www.hunterhacks.com/",
				Status:   shared.StatusPosted,
				StartISO: "2025-09-15 17:00:00",
				EndISO:   "2025-09-15 19:00:00",
			},
			{
				ID:       2,
				Title:    "cs workshop",
				Location: "Room 304, narnia",
				RSVPLink: "https://www.hunterhacks.com/",
				Status:   shared.StatusPosted,
				StartISO: "2025-10-28 11:00:00",
				EndISO:   "2025-10-28 19:00:00",
			},
			{
				ID:       2,
				Title:    "cooking workshop",
				Location: "Room 505, Hogwarts",
				RSVPLink: "https://www.hunterhacks.com/",
				Status:   shared.StatusPosted,
				StartISO: "2025-11-28 10:00:00",
				EndISO:   "2025-11-28 13:00:00",
			},
		},
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
