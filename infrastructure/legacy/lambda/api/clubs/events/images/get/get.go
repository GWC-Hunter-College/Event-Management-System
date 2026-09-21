package main

import (
	"cdk-infrastructure/database/models"
	"cdk-infrastructure/utils/query_client"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client *s3.Client
	bucket   string

	queryClient *query_client.QueryClient
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	s3Client = s3.NewFromConfig(cfg)

	bucketVar, ok := os.LookupEnv("S3_BUCKET")

	if !ok {
		panic("S3_BUCKET environment variable not set")
	}

	bucket = bucketVar

	secretArn := os.Getenv("DB_SECRET_ARN")
	host := os.Getenv("DB_HOST")
	database := os.Getenv("DB_NAME")

	queryClientVar, err := query_client.NewClientFromHost(context.Background(), secretArn, database, host)
	if err != nil {
		log.Printf("Could not create query client: %v", err)

		panic("Could not create query client: " + err.Error())
	}

	queryClient = queryClientVar
}

type ResponseImage struct {
	ImageID   string `json:"imageId"`
	Type      string `json:"mimetype"`
	SourceUrl string `json:"sourceUrl"`
}

type ResponseJSONSchema struct {
	Images []ResponseImage `json:"images"`
}

type SelectQuerySchema struct {
	EventID string       `db:"fk_event_id"  json:"eventId"`
	Image   models.Image `db:"*"            json:"image"`
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	defer queryClient.Conn.Close()

	eventId := request.PathParameters["eventId"]

	var eventImages = []SelectQuerySchema{}
	selectEventImagesQuery := query_client.NewQuery("images/SELECT_event_images.sql", eventId)
	err := queryClient.Select(&eventImages, selectEventImagesQuery)

	if err != nil {
		log.Printf("Error retrieving images: %v", err)

		response, _ := json.Marshal(map[string]string{
			"message": "Could not retrieve images: " + err.Error(),
		})

		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       string(response),
		}, nil
	}

	presignClient := s3.NewPresignClient(s3Client)

	outputImages := []ResponseImage{}

	for _, eventImage := range eventImages {

		command := &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(eventImage.Image.ObjectKey),
		}

		presignedGet, err := presignClient.PresignGetObject(ctx, command, func(o *s3.PresignOptions) {
			o.Expires = time.Minute
		})

		if err != nil {
			response, _ := json.Marshal(map[string]string{
				"message": "Could not generate presigned URL: " + err.Error(),
			})

			return events.APIGatewayV2HTTPResponse{
				StatusCode: 500,
				Body:       string(response),
			}, nil
		}

		outputImages = append(outputImages, ResponseImage{
			ImageID:   eventImage.Image.ImageID,
			Type:      eventImage.Image.Type,
			SourceUrl: presignedGet.URL,
		})
	}

	jsonResponse, _ := json.Marshal(ResponseJSONSchema{
		Images: outputImages,
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
			"Access-Control-Allow-Headers": "*",
		},
		Body: string(jsonResponse),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
