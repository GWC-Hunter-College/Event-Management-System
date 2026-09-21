package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	s3Client *s3.Client
	bucket   string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	s3Client = s3.NewFromConfig(cfg)

	bucketVar, ok := os.LookupEnv("S3_BUCKET")

	if !ok {
		panic("S3_BUCKET environment variable not set")
	}

	bucket = bucketVar
}

type S3Response struct {
	UploadURL string `json:"uploadUrl"`
	ImageID   string `json:"imageId"`
	ObjectKey string `json:"objectKey"`
}

type RequestBody struct {
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	eventId := request.PathParameters["eventId"]

	var body RequestBody
	err := json.Unmarshal([]byte(request.Body), &body)

	if err != nil || body.Filename == "" || body.Mimetype == "" {
		return events.APIGatewayV2HTTPResponse{
			Body: string(`message: "Missing filename or mimetype"`), StatusCode: 400,
		}, nil
	}

	imageExtension := filepath.Ext(body.Filename)
	imageUUID := uuid.NewString()

	objectKey := fmt.Sprintf("events/%s/thumbnails/%s%s", eventId, imageUUID, imageExtension)

	command := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(body.Mimetype),
	}

	presignClient := s3.NewPresignClient(s3Client)

	presignedPut, err := presignClient.PresignPutObject(ctx, command, func(o *s3.PresignOptions) {
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

	jsonResponse, _ := json.Marshal(S3Response{
		UploadURL: presignedPut.URL,
		ImageID:   imageUUID,
		ObjectKey: objectKey,
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
