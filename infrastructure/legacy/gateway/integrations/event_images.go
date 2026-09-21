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

// Integrations for event image presign endpoint
func GetEventImagesIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	s3Params gateway_parameters.S3PermissionsParameters,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	bucket := s3Params.Bucket

	dbSecretArn := dbParams.Secret.SecretArn()
	dbInstance := dbParams.DbInstance
	dbSecret := dbParams.Secret
	dbHost := jsii.String(dbParams.DbHost)
	dbName := jsii.String(dbParams.DbName)

	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetEventImagesPresignFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubEventImagesPresign" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/events/images/get/get.go"),
		Environment: &map[string]*string{
			"S3_BUCKET":     bucket.BucketName(),
			"DB_SECRET_ARN": dbSecretArn,
			"DB_HOST":       dbHost,
			"DB_NAME":       dbName,
		},
		Vpc: vpc,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
			lambdaToSecretsSG,
		},
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
	})

	gateway_helpers.GrantRdsAccessToLambda(function, dbInstance, dbSecret)
	gateway_helpers.GrantS3AccessToLambda(function, bucket, "events/*", true, false)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetEventImagePresignIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

func PostEventImagesIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	s3Params gateway_parameters.S3PermissionsParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	bucket := s3Params.Bucket

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostEventImagesPresignFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostClubEventImagesPresign" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/events/images/post/post.go"),
		Environment: &map[string]*string{
			"S3_BUCKET": bucket.BucketName(),
		},
		Vpc: vpc,
	})

	gateway_helpers.GrantS3AccessToLambda(function, bucket, "events/*", false, true)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("PostEventImagePresignIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

func ConfirmEventImagesIntegration(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	dbSecretArn := dbParams.Secret.SecretArn()
	dbHost := jsii.String(dbParams.DbHost)
	dbName := jsii.String(dbParams.DbName)

	dbInstance := dbParams.DbInstance
	dbSecret := dbParams.Secret

	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("ConfirmEventImagesFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("ConfirmClubEventImages" + deploymentTarget),
		Entry:        jsii.String("lambda/api/clubs/events/images/confirm/post.go"),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": dbSecretArn,
			"DB_HOST":       dbHost,
			"DB_NAME":       dbName,
		},
		Vpc: vpc,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
			lambdaToSecretsSG,
		},
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
	})

	gateway_helpers.GrantRdsAccessToLambda(function, dbInstance, dbSecret)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("ConfirmEventImageIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}
