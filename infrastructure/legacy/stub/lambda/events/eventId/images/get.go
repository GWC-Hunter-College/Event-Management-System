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

	resp := shared.Images{
		Images: []shared.Image{
			{
				ID:        1,
				Purpose:   fmt.Sprintf("Image 1 of Event %s", eventID),
				URL:       "https://bit.ly/fcc-relaxing-cat",
				CreateISO: "2025-09-05 19:00:00",
			},
			{
				ID:        2,
				Purpose:   fmt.Sprintf("Image 2 of Event %s", eventID),
				URL:       "https://bit.ly/fcc-relaxing-cat",
				CreateISO: "2025-09-05 19:00:00",
			},
			{
				ID:        3,
				Purpose:   fmt.Sprintf("Image 3 of Event %s", eventID),
				URL:       "https://bit.ly/fcc-relaxing-cat",
				CreateISO: "2025-09-05 19:00:00",
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
