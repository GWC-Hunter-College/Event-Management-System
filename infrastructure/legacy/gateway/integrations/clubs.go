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

// Integration for GET /clubs?verified=true
func GetClubsByVerification(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetClubsByVerificationFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubsByVerification" + deploymentTarget),
		Description:  jsii.String("Get only verified clubs if verified is true, else get all clubs"),
		Entry:        jsii.String("lambda/api/clubs/get.go"),
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
		jsii.String("GetClubsByVerificationIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for GET /clubs/{clubId}
func GetClubById(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetClubByIdFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubsById" + deploymentTarget),
		Description:  jsii.String("Get by its clubId"),
		Entry:        jsii.String("lambda/api/clubs/clubId/get.go"),
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
		jsii.String("GetClubByIdIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for GET /clubs/{clubId}/events
func GetClubEventsByPeriod(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("GetClubEventsByPeriodFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("GetClubEventsByPeriod" + deploymentTarget),
		Description:  jsii.String("Get posted events from a specific club between start and end date"),
		Entry:        jsii.String("lambda/api/clubs/clubId/events/get.go"),
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
		jsii.String("GetClubEventsByPeriodIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

// Integration for POST /clubs/{clubId}/events
func PostNewClubEvent(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostNewClubEventFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostNewClubEvent" + deploymentTarget),
		Description:  jsii.String("Post a new event to a specific club"),
		Entry:        jsii.String("lambda/api/clubs/clubId/events/post/post.go"),
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
		jsii.String("PostNewClubEventIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

func PostNewClub(
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

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("CreateNewClubFunction"+deploymentTarget), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("CreateNewClubFunction" + deploymentTarget),
		Description:  jsii.String("Create a new club"),
		Entry:        jsii.String("lambda/api/clubs/post/post.go"),
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
		jsii.String("CreateNewClubIntegration"+deploymentTarget),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}

func PostJoinClubMemberMe(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	host := dbParams.DbHost
	dbName := dbParams.DbName
	arn := dbParams.Secret.SecretArn()
	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsManagerSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(
		stack,
		jsii.String("PostJoinClubMemberMeFunction"),
		&awscdklambdagoalpha.GoFunctionProps{
			FunctionName: jsii.String("PostJoinClubMemberMe"),
			Description:  jsii.String("Join a club as the current student"),
			Entry:        jsii.String("lambda/api/clubs/clubId/members/me/post/post.go"),
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
		},
	)

	gateway_helpers.GrantRdsAccessToLambda(function, dbParams.DbInstance, dbParams.Secret)

	return awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("PostJoinClubMemberMeIntegration"),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)
}

func DeleteClubMemberMe(
	stack awscdk.Stack,
	vpc awsec2.IVpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
) awsapigatewayv2integrations.HttpLambdaIntegration {
	host := dbParams.DbHost
	dbName := dbParams.DbName
	arn := dbParams.Secret.SecretArn()
	lambdaSG := dbParams.LambdaSG
	lambdaToSecretsManagerSG := dbParams.LambdaToSecretsManagerSG

	function := awscdklambdagoalpha.NewGoFunction(
		stack,
		jsii.String("DeleteClubMemberMeFunction"),
		&awscdklambdagoalpha.GoFunctionProps{
			FunctionName: jsii.String("DeleteClubMemberMe"),
			Description:  jsii.String("Leave a club as the current student"),
			Entry:        jsii.String("lambda/api/clubs/clubId/members/me/delete/delete.go"),
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
		},
	)

	gateway_helpers.GrantRdsAccessToLambda(function, dbParams.DbInstance, dbParams.Secret)

	return awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("DeleteClubMemberMeIntegration"),
		function,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)
}
