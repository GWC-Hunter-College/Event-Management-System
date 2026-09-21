package integrations

import (
	gateway_parameters "cdk-infrastructure/gateway/parameters"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"
)

// Integrations for club thumbnail presign endpoint
func GetClubThumbnailsIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	lambdaToProxySG awsec2.ISecurityGroup,
	s3Params gateway_parameters.S3PermissionsParameters,
	deploymentTarget string,
) (awsapigatewayv2integrations.HttpLambdaIntegration, awscdklambdagoalpha.GoFunction) {
	bucket := s3Params.Bucket

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetClubThumbnailFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubThumbnailsPresign" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/thumbnails/get/get.go"),
		Environment: &map[string]*string{
			"S3_BUCKET": bucket.BucketName(),
		},
		Vpc: vpc,
	})

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetClubThumbnailPresignIntegration"),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration, function
}

func PostClubThumbnailsIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	s3Params gateway_parameters.S3PermissionsParameters,
	deploymentTarget string,
) (awsapigatewayv2integrations.HttpLambdaIntegration, awscdklambdagoalpha.GoFunction) {
	bucket := s3Params.Bucket

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostClubThumbnailPresignFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostClubThumbnailPresign" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/thumbnails/post/post.go"),
		Environment: &map[string]*string{
			"S3_BUCKET": bucket.BucketName(),
		},
		Vpc: vpc,
	})

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("PostClubThumbnailPresignIntegration"),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration, function
}
