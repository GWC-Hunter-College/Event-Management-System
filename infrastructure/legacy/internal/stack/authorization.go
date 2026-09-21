package stack

import (
	"os"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2authorizers"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	gateway_helpers "cdk-infrastructure/gateway/helpers"

	cr "github.com/aws/aws-cdk-go/awscdk/v2/customresources"
)

type AuthorizationStackProps struct {
	Props awscdk.StackProps

	Vpc                               awsec2.Vpc
	LambdaSecretsManagerSecurityGroup awsec2.SecurityGroup
	LambdaSecurityGroup               awsec2.SecurityGroup
	DbInstance                        awsrds.DatabaseInstance

	UserPool  awscognito.IUserPool
	AppClient awscognito.IUserPoolClient
}

type AuthorizationStack struct {
	Stack      awscdk.Stack
	Authorizer awsapigatewayv2.IHttpRouteAuthorizer
}

func NewAuthorizationStack(scope constructs.Construct, id string, props *AuthorizationStackProps) *AuthorizationStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	//  =======================================
	//  read props
	//  =======================================
	vpc := props.Vpc
	dbInstance := props.DbInstance

	lambdaSecretsManagerSecurityGroup := props.LambdaSecretsManagerSecurityGroup
	lambdaSecurityGroup := props.LambdaSecurityGroup

	//  =======================================
	//  grab prod status
	//  =======================================
	var databaseName string
	productionStatus := strings.ToLower(os.Getenv("PRODUCTION_STATUS"))

	if productionStatus == "true" {
		databaseName = "PRODUCTION"
	} else {
		databaseName = "STAGING"
	}

	//  =======================================
	//  authenticaion lambda
	//  =======================================

	postConfirmFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("PostConfirmUserUpsertFunction"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("PostConfirmUserUpsert"),
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
		SourceArn: props.UserPool.UserPoolArn(),
	})

	physID := "Wire-" + *props.UserPool.UserPoolId()

	onUp := &cr.AwsSdkCall{
		Service: jsii.String("CognitoIdentityServiceProvider"), // <<< v2 name
		Action:  jsii.String("updateUserPool"),
		Parameters: &map[string]interface{}{
			"UserPoolId": *props.UserPool.UserPoolId(),
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
			"UserPoolId":   *props.UserPool.UserPoolId(),
			"LambdaConfig": map[string]interface{}{}, // clears triggers
		},
		PhysicalResourceId:       cr.PhysicalResourceId_Of(jsii.String(physID)),
		IgnoreErrorCodesMatching: jsii.String(".*ResourceNotFound.*"),
	}

	// Scope permissions if you want (instead of ANY_RESOURCE)
	policy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   &[]*string{jsii.String("cognito-idp:UpdateUserPool")},
		Resources: &[]*string{props.UserPool.UserPoolArn()},
	})

	cr.NewAwsCustomResource(stack, jsii.String("WireCognitoTriggers"), &cr.AwsCustomResourceProps{
		Policy:              cr.AwsCustomResourcePolicy_FromStatements(&[]awsiam.PolicyStatement{policy}),
		OnCreate:            onUp,
		OnUpdate:            onUp,
		OnDelete:            onDel,
		InstallLatestAwsSdk: jsii.Bool(false), // use SDK v2
	})

	// authorizer
	// authorizer := apigwauth.NewHttpUserPoolAuthorizer(
	// 	jsii.String("CognitoAuth"),
	// 	props.UserPool,
	// 	&apigwauth.HttpUserPoolAuthorizerProps{
	// 		UserPoolClients: &[]awscognito.IUserPoolClient{props.AppClient},
	// 	},
	// )

	// authMeFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("Auth Me Function"), &awscdklambdagoalpha.GoFunctionProps{
	// 	FunctionName: jsii.String("AuthMeFunction"),
	// 	Entry:        jsii.String("./lambda/auth/me/main.go"),
	// })

	authorizer := awsapigatewayv2authorizers.NewHttpUserPoolAuthorizer(
		jsii.String("CognitoJwt"),
		props.UserPool,
		&awsapigatewayv2authorizers.HttpUserPoolAuthorizerProps{
			UserPoolClients: &[]awscognito.IUserPoolClient{props.AppClient},
			// Optional: AuthorizerName: jsii.String("CognitoJwtAuthorizer"),
			// Optional: ResultsCacheTtl: awscdk.Duration_Seconds(jsii.Number(0)), // while developing
		},
	)

	return &AuthorizationStack{
		Stack:      stack,
		Authorizer: authorizer,
	}
}
