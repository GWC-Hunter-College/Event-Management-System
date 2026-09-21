package main

import (
	"cdk-infrastructure/utils/query_client"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	queryClient *query_client.QueryClient
)

func init() {
	secretArn := os.Getenv("DB_SECRET_ARN")
	host := os.Getenv("DB_HOST")
	database := os.Getenv("DB_NAME")

	query_client, err := query_client.NewClientFromHost(context.Background(), secretArn, database, host)
	if err != nil {
		panic("Could not create query client: " + err.Error())
	}

	queryClient = query_client
}

type ResponseJSONSchema struct {
	Message string `json:"message"`
}

type RequestBodySchema struct {
	ImageID   string `json:"imageId"`
	ObjectKey string `json:"objectKey"`
	Filename  string `json:"filename"`
	Mimetype  string `json:"mimetype"`
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	defer queryClient.Close()

	eventId := request.PathParameters["eventId"]

	var body RequestBodySchema
	err := json.Unmarshal([]byte(request.Body), &body)

	if err != nil || body.ImageID == "" || body.ObjectKey == "" || body.Filename == "" || body.Mimetype == "" {
		return events.APIGatewayV2HTTPResponse{
			Body:       string(`message: "Missing imageId, objectKey, filename, or mimetype in request body"`),
			StatusCode: 400,
		}, nil
	}

	insertImageQuery := query_client.NewQuery("images/INSERT_image.sql",
		body.ImageID, "event-image", body.ObjectKey, body.Filename, body.Mimetype,
	)

	insertImageResult, err := queryClient.Exec(insertImageQuery)

	if err != nil {

		log.Printf("Error inserting image: %v", err)

		return events.APIGatewayV2HTTPResponse{
			Body:       string(`message: "Database error: ` + err.Error() + `"`),
			StatusCode: 500,
		}, nil
	}

	log.Printf("Inserted %d images", insertImageResult)

	insertEventImageQuery := query_client.NewQuery("images/INSERT_event_image.sql",
		eventId, body.ImageID,
	)

	insertEventImageResult, err := queryClient.Exec(insertEventImageQuery)

	if err != nil {
		log.Printf("Error inserting event image: %v", err)

		return events.APIGatewayV2HTTPResponse{
			Body:       string(`message: "Database error: ` + err.Error() + `"`),
			StatusCode: 500,
		}, nil
	}

	log.Printf("Inserted %d event images", insertEventImageResult)

	response := ResponseJSONSchema{
		Message: "Image confirmed successfully",
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)

		return events.APIGatewayV2HTTPResponse{
			Body:       string(`message: "Internal JSON marshalling error"` + err.Error()),
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		Body:       string(responseJson),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
