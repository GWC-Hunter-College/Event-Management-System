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

// Integration for GET /me
func GetStudentMeBySub(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	host := dbParams.DbHost
	dbName := dbParams.DbName
	arn := dbParams.Secret.SecretArn()
	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsManagerSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetStudentMeFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetStudentMe" + deploymentTarget),
		Description:  jsii.String("Returns the student data associated with the sub from jwt"),
		Entry:        jsii.String("./lambda/api/me/get.go"),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": arn,
			"DB_HOST":       jsii.String(host),
			"DB_NAME":       jsii.String(dbName),
		},
		Vpc:     vpc,
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
			lambdaToSecretsManagerSG,
		},
	})

	gateway_helpers.GrantRdsAccessToLambda(function, dbParams.DbInstance, dbParams.Secret)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetStudentMeIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for GET /me/clubs
func GetMyClubs(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	host := dbParams.DbHost
	dbName := dbParams.DbName
	arn := dbParams.Secret.SecretArn()
	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsManagerSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetMyClubsFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetMyClubs" + deploymentTarget),
		Description:  jsii.String("Returns clubs of the student of associated jwt"),
		Entry:        jsii.String("lambda/api/me/clubs/get.go"),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": arn,
			"DB_HOST":       jsii.String(host),
			"DB_NAME":       jsii.String(dbName),
		},
		Vpc:     vpc,
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
			lambdaToSecretsManagerSG,
		},
	})

	gateway_helpers.GrantRdsAccessToLambda(function, dbParams.DbInstance, dbParams.Secret)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetMyClubsIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for GET /me/clubs
func GetMyEvents(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	host := dbParams.DbHost
	dbName := dbParams.DbName
	arn := dbParams.Secret.SecretArn()
	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsManagerSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetMyEventsFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetMyEvents" + deploymentTarget),
		Description:  jsii.String("Returns events from clubs that a user has joined"),
		Entry:        jsii.String("lambda/api/me/events/get.go"),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": arn,
			"DB_HOST":       jsii.String(host),
			"DB_NAME":       jsii.String(dbName),
		},
		Vpc:     vpc,
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSG,
			lambdaToSecretsManagerSG,
		},
	})

	gateway_helpers.GrantRdsAccessToLambda(function, dbParams.DbInstance, dbParams.Secret)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("GetMyEventsIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}
