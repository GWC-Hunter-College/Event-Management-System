package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"cdk-infrastructure/stub/shared"
)

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	v := strings.ToLower(req.QueryStringParameters["verified"])
	onlyVerified := v == "true" || v == "1" || v == "yes"

	verified := []shared.Club{
		{ID: 1, Name: "CS Club"},
		{ID: 2, Name: "Girls Who Code"},
		{ID: 2, Name: "Wics"},
	}
	unverified := []shared.Club{
		{ID: 3, Name: "Cooking Club"},
	}

	var out []shared.Club
	if onlyVerified {
		out = verified
	} else {
		out = append([]shared.Club{}, verified...)
		out = append(out, unverified...)
	}

	resp := shared.Clubs{Clubs: out}

	b, _ := json.Marshal(resp)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

func main() { lambda.Start(handle) }
