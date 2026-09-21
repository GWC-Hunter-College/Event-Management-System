package stack

import (
	gateway_helpers "cdk-infrastructure/gateway/helpers"
	gateway_parameters "cdk-infrastructure/gateway/parameters"
	gateway_routes "cdk-infrastructure/gateway/routes"

	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2authorizers"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	cr "github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DevApiStackProps struct {
	Props awscdk.StackProps

	Vpc                               awsec2.Vpc
	LambdaSecretsManagerSecurityGroup awsec2.SecurityGroup
	LambdaSecurityGroup               awsec2.SecurityGroup
	DbInstance                        awsrds.DatabaseInstance

	BucketName *string
	Bucket     awss3.Bucket

	UserPool  awscognito.IUserPool
	AppClient awscognito.IUserPoolClient
}

func NewDevApiStack(scope constructs.Construct, id string, props *DevApiStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("ClubEventApiDev"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("ClubEventApiDev"),
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowHeaders: &[]*string{
				jsii.String("*"),
			},
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_GET,
				awsapigatewayv2.CorsHttpMethod_POST,
				awsapigatewayv2.CorsHttpMethod_OPTIONS,
				awsapigatewayv2.CorsHttpMethod_PATCH,
			},
			AllowOrigins: &[]*string{
				jsii.String("*"), // allowing from all origins atm, should be locked down later
			},
		},
	})

	const databaseName string = "STAGING"

	vpc := props.Vpc

	authorizer := createDevApiAuthorizer(
		stack,
		props.Vpc,
		props.LambdaSecurityGroup,
		props.LambdaSecretsManagerSecurityGroup,
		props.DbInstance,
		props.UserPool,
		props.AppClient,
		databaseName,
	)

	s3Params := gateway_parameters.S3PermissionsParameters{
		Bucket: props.Bucket,
	}

	dbParams := gateway_parameters.DatabaseConnectionParameters{
		Secret:                   props.DbInstance.Secret(),
		LambdaSG:                 props.LambdaSecurityGroup,
		LambdaToSecretsManagerSG: props.LambdaSecretsManagerSecurityGroup,
		DbInstance:               props.DbInstance,
		DbHost:                   *props.DbInstance.DbInstanceEndpointAddress(),
		DbName:                   databaseName,
	}

	deploymentTarget := "Dev"

	gateway_routes.TestRoutes(httpApi, stack, deploymentTarget)

	gateway_routes.ClubRoutes(httpApi, stack, vpc, dbParams, authorizer, deploymentTarget)

	gateway_routes.ClubImageRoutes(httpApi, stack, vpc, s3Params, dbParams, deploymentTarget)

	gateway_routes.EventImageRoutes(httpApi, stack, vpc, s3Params, dbParams, deploymentTarget)

	gateway_routes.PublicEventRoutes(httpApi, stack, vpc, dbParams, deploymentTarget)

	gateway_routes.StudentMeRoutes(httpApi, stack, vpc, dbParams, authorizer, deploymentTarget)

	// log HTTP API endpoint
	awscdk.NewCfnOutput(stack, jsii.String("myHttpApiEndpoint"), &awscdk.CfnOutputProps{
		Value:       httpApi.ApiEndpoint(),
		Description: jsii.String("HTTP API Endpoint"),
	})

	return stack
}

func createDevApiAuthorizer(
	stack awscdk.Stack,
	vpc awsec2.Vpc,
	lambdaSecurityGroup awsec2.SecurityGroup,
	lambdaSecretsManagerSecurityGroup awsec2.SecurityGroup,
	dbInstance awsrds.DatabaseInstance,
	userPool awscognito.IUserPool,
	appClient awscognito.IUserPoolClient,
	databaseName string,
) awsapigatewayv2.IHttpRouteAuthorizer {
	postConfirmFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostConfirmUserUpsertFunctionDev"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostConfirmUserUpsertDev"),
		Description:  jsii.String("Upsert Cognito User to DB"),
		Entry:        jsii.String("./lambda/internal/auth/postConfirm/upsert.go"),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": dbInstance.Secret().SecretArn(),
			"DB_HOST":       jsii.String(*dbInstance.DbInstanceEndpointAddress()),
			"DB_NAME":       jsii.String(databaseName),
		},
		Vpc:     vpc,
		Timeout: awscdk.Duration_Minutes(jsii.Number(1)),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
			lambdaSecretsManagerSecurityGroup,
		},
	})

	gateway_helpers.GrantRdsAccessToLambda(postConfirmFunction, dbInstance, dbInstance.Secret())

	// allow Cognito to invoke your Lambda
	postConfirmFunction.AddPermission(jsii.String("AllowCognitoInvoke"), &awslambda.Permission{
		Principal: awsiam.NewServicePrincipal(jsii.String("cognito-idp.amazonaws.com"), nil),
		SourceArn: userPool.UserPoolArn(),
	})

	physID := "Wire-" + *userPool.UserPoolId()

	onUp := &cr.AwsSdkCall{
		Service: jsii.String("CognitoIdentityServiceProvider"), // <<< v2 name
		Action:  jsii.String("updateUserPool"),
		Parameters: &map[string]interface{}{
			"UserPoolId": *userPool.UserPoolId(),
			"LambdaConfig": map[string]interface{}{
				"PostConfirmation": postConfirmFunction.FunctionArn(),
				// include if you want every login too:
				"PostAuthentication": postConfirmFunction.FunctionArn(),
			},
		},
		PhysicalResourceId: cr.PhysicalResourceId_Of(jsii.String(physID)),
	}

	onDel := &cr.AwsSdkCall{
		Service: jsii.String("CognitoIdentityServiceProvider"), // <<< v2 name
		Action:  jsii.String("updateUserPool"),
		Parameters: &map[string]interface{}{
			"UserPoolId":   *userPool.UserPoolId(),
			"LambdaConfig": map[string]interface{}{}, // clears triggers
		},
		PhysicalResourceId:       cr.PhysicalResourceId_Of(jsii.String(physID)),
		IgnoreErrorCodesMatching: jsii.String(".*ResourceNotFound.*"),
	}

	// Scope permissions if you want (instead of ANY_RESOURCE)
	policy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   &[]*string{jsii.String("cognito-idp:UpdateUserPool")},
		Resources: &[]*string{userPool.UserPoolArn()},
	})

	cr.NewAwsCustomResource(stack, jsii.String("WireCognitoTriggers"), &cr.AwsCustomResourceProps{
		Policy:              cr.AwsCustomResourcePolicy_FromStatements(&[]awsiam.PolicyStatement{policy}),
		OnCreate:            onUp,
		OnUpdate:            onUp,
		OnDelete:            onDel,
		InstallLatestAwsSdk: jsii.Bool(false), // use SDK v2
	})

	authorizer := awsapigatewayv2authorizers.NewHttpUserPoolAuthorizer(
		jsii.String("CognitoJwtDev"),
		userPool,
		&awsapigatewayv2authorizers.HttpUserPoolAuthorizerProps{
			UserPoolClients: &[]awscognito.IUserPoolClient{appClient},
			// Optional: AuthorizerName: jsii.String("CognitoJwtAuthorizer"),
			// Optional: ResultsCacheTtl: awscdk.Duration_Seconds(jsii.Number(0)), // while developing
		},
	)

	return authorizer
}
