package integrations

import (
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	gateway_parameters "cdk-infrastructure/gateway/parameters"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"
)

// Integrations for event thumbnail presign endpoint
func GetEventThumbnailPresignFunction(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	lambdaToProxySG awsec2.ISecurityGroup,
	s3Params gateway_parameters.S3PermissionsParameters,
	dbSecretArn *string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	bucket := s3Params.Bucket

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetEventThumbnailPresignFunction"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubEventThumbnailPresign"),
		Entry:        jsii.String("lambda/api/clubs/events/thumbnails/get/get.go"),
		Environment: &map[string]*string{
			"S3_BUCKET":     bucket.BucketName(),
			"DB_SECRET_ARN": dbSecretArn,
		},
		Vpc: vpc,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaToProxySG,
		},
	})

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetEventThumbnailPresignIntegration"),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

func PostEventThumbnailsIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	s3Params gateway_parameters.S3PermissionsParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	bucket := s3Params.Bucket

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostEventThumbnailsPresignFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostClubEventThumbnailsPostPresign" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/events/thumbnails/post/post.go"),
		Environment: &map[string]*string{
			"S3_BUCKET": bucket.BucketName(),
		},
		Vpc: vpc,
	})

	gateway_helpers.GrantS3AccessToLambda(function, bucket, "events/*/thumbnails/*", false, true)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("PostEventThumbnailsPresignIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}
