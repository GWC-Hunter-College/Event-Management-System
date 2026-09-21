package gateway_routes

import (
	"cdk-infrastructure/gateway/integrations"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/jsii-runtime-go"
)

// Helper function that sets up routes for testing purposes, such as health checks.
// Routes:
//
// - GET /health: A simple health check endpoint that returns a "Server Running" response.
func TestRoutes(httpApi awsapigatewayv2.HttpApi, stack awscdk.Stack, deploymentTarget string) {
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/health"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integrations.PingTestIntegration(stack, deploymentTarget),
	})
}
