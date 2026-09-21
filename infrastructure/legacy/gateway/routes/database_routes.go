package gateway_routes

import (
	"cdk-infrastructure/gateway/integrations"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/jsii-runtime-go"
)

type DatabaseRouteProps struct {
	Vpc                               awsec2.Vpc
	LambdaSecretsManagerSecurityGroup awsec2.SecurityGroup
	DbInstance                        awsrds.DatabaseInstance
	ProxyEndpoint                     *string
	LambdaSecurityGroup               awsec2.SecurityGroup
}

// Helper function to add database-related routes to the HTTP API
// Routes:
//
//	GET /database/test
func DatabaseRoutes(httpApi awsapigatewayv2.HttpApi, stack awscdk.Stack, props DatabaseRouteProps) {
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:    jsii.String("/database/test"),
		Methods: &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integrations.DatabaseTestIntegration(stack, integrations.DatabaseTestIntegrationProps{
			Vpc:                               props.Vpc,
			LambdaSecretsManagerSecurityGroup: props.LambdaSecretsManagerSecurityGroup,
			LambdaSecurityGroup:               props.LambdaSecurityGroup,
			DbInstance:                        props.DbInstance,
			ProxyEndpoint:                     props.ProxyEndpoint,
		}),
	})
}
