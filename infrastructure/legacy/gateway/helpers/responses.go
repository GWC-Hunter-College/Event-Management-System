package gateway_helpers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// NewSuccessGatewayResponse constructs an API Gateway success response with status code 200 and the supplied body.
//
// Response body is constructed as below:
//
//	{
//	    "message": "successMessage",
//	    ...extraFields
//	}
func NewSuccessGatewayResponse(message string, extraFields map[string]any) (events.APIGatewayV2HTTPResponse, error) {
	response := map[string]any{
		"message": message,
	}

	for k, v := range extraFields {
		response[k] = v
	}

	body, err := json.Marshal(response)

	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error - Marshalling failed: " + err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}, nil
}

// NewErrorGatewayResponse constructs an API Gateway error response with status code 400 and the supplied body.
//
// Response body is constructed as below:
//
//	{
//	    "error": "errorMessage",
//	    ...extraFields
//	}
func NewClientErrorGatewayResponse(errorMessage string, extraFields map[string]any) (events.APIGatewayV2HTTPResponse, error) {
	response := map[string]any{
		"error": errorMessage,
	}

	for k, v := range extraFields {
		response[k] = v
	}

	body, err := json.Marshal(response)

	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error - Marshalling failed: " + err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 400,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}, nil
}

// NewErrorGatewayResponse constructs an API Gateway error response with status code 404 and the supplied body.
//
// Response body is constructed as below:
//
//	{
//	    "error": "errorMessage",
//	    ...extraFields
//	}
func NewNotFoundGatewayResponse(errorMessage string, extraFields map[string]any) (events.APIGatewayV2HTTPResponse, error) {
	response := map[string]any{
		"error": errorMessage,
	}

	for k, v := range extraFields {
		response[k] = v
	}

	body, err := json.Marshal(response)

	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error - Marshalling failed: " + err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}, nil
}

// NewErrorGatewayResponse constructs an API Gateway error response with status code 500 and the supplied body.
//
// Response body is constructed as below:
//
//	{
//	    "error": "errorMessage",
//	    ...extraFields
//	}
func NewServerErrorGatewayResponse(errorMessage string, extraFields map[string]any) (events.APIGatewayV2HTTPResponse, error) {
	response := map[string]any{
		"error": errorMessage,
	}

	for k, v := range extraFields {
		response[k] = v
	}

	body, err := json.Marshal(response)

	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error - Marshalling failed: " + err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 500,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}, nil
}
