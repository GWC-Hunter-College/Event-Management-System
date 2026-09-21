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

// Integration for GET /events?startDate={startDate}&endDate={endDate}
func GetEventsByPeriod(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetEventsByPeriodFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetEventsByPeriod" + deploymentTarget),
		Description:  jsii.String("Get posted events between start and end date"),
		Entry:        jsii.String("lambda/api/events/get.go"),
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
		jsii.String("GetEventsByPeriodIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for GET /events/{eventId}
func GetEventById(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetEventByIdFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetEventById" + deploymentTarget),
		Description:  jsii.String("Get event by ID"),
		Entry:        jsii.String("lambda/api/events/eventId/get.go"),
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
		jsii.String("GetEventByIdIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}
