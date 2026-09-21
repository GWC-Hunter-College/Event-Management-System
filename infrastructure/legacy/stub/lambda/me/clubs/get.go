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
	resp := shared.Clubs{
		Clubs: []shared.Club{
			{
				ID:           2,
				Name:         "Girls Who Code @ Hunter",
				Role:         shared.RoleEboard,
				ThumbnailURL: "https://ugc.production.linktr.ee/hRxlRoDSWCCPkrC2fYb5_XY24EQSCVo7Z3yP6?io=true&size=avatar-v3_0",
			},
			{
				ID:           3,
				Name:         "Hunter CS Club",
				Role:         shared.RoleOwner,
				ThumbnailURL: "https://ugc.production.linktr.ee/hRxlRoDSWCCPkrC2fYb5_XY24EQSCVo7Z3yP6?io=true&size=avatar-v3_0",
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
