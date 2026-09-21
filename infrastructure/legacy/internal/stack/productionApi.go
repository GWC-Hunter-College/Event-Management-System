package stack

import (
	gateway_parameters "cdk-infrastructure/gateway/parameters"
	gateway_routes "cdk-infrastructure/gateway/routes"

	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ProdApiStackProps struct {
	Props awscdk.StackProps

	Vpc                               awsec2.Vpc
	LambdaSecretsManagerSecurityGroup awsec2.SecurityGroup
	LambdaSecurityGroup               awsec2.SecurityGroup
	DbInstance                        awsrds.DatabaseInstance

	BucketName *string
	Bucket     awss3.Bucket

	Authorizer awsapigatewayv2.IHttpRouteAuthorizer
}

func NewProdApiStack(scope constructs.Construct, id string, props *ProdApiStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("ClubEventApiProd"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("ClubEventApiProd"),
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowHeaders: &[]*string{
				jsii.String("*"),
			},
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_GET,
				awsapigatewayv2.CorsHttpMethod_POST,
				awsapigatewayv2.CorsHttpMethod_OPTIONS,
				awsapigatewayv2.CorsHttpMethod_PATCH,
				awsapigatewayv2.CorsHttpMethod_DELETE,
			},
			AllowOrigins: &[]*string{
				jsii.String("*"), // allowing from all origins atm, should be locked down later
			},
		},
	})

	const databaseName string = "PRODUCTION"

	vpc := props.Vpc

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

	deploymentTarget := "Prod"

	authorizer := props.Authorizer

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
