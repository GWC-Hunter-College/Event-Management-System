package gateway_routes

import (
	"cdk-infrastructure/gateway/integrations"
	gateway_parameters "cdk-infrastructure/gateway/parameters"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
)

// Helper function to add public event routes to the API.
//
// Routes:
//
// GET /events?startDate={startDate}&endDate={endDate}
// GET /events/{eventId}
func PublicEventRoutes(
	httpApi awsapigatewayv2.HttpApi,
	stack awscdk.Stack,
	vpc awsec2.Vpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) {
	// /events?startDate={startDate}&endDate={endDate}
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/events"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
		Integration: integrations.GetEventsByPeriod(stack, vpc, dbParams, deploymentTarget),
	})

	// /events/{eventId}
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/events/{eventId}"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
		Integration: integrations.GetEventById(stack, vpc, dbParams, deploymentTarget),
	})
}
