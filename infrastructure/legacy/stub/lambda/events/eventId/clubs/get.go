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
	eventID := req.PathParameters["eventId"]
	_ = eventID

	resp := shared.EventOwners{
		Owner: shared.Club{
			ID:   2,
			Name: "Girls Who Code @ Hunter",
		},
		Associates: []shared.Club{
			{ID: 3, Name: "Hunter CS Club"},
			{ID: 3, Name: "Wics"},
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
