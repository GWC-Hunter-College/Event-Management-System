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

	resp := shared.EventDescription{
		Description: fmt.Sprintf("Description for the event %s", eventID),
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
