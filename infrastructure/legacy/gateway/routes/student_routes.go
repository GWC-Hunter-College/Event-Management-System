package gateway_routes

import (
	"cdk-infrastructure/gateway/integrations"
	gateway_parameters "cdk-infrastructure/gateway/parameters"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
)

// Helper function to add club routes to the API.
// Routes:
//
// POST /clubs/{clubId}/events
func StudentMeRoutes(
	httpApi awsapigatewayv2.HttpApi,
	stack awscdk.Stack,
	vpc awsec2.Vpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	Authorizer awsapigatewayv2.IHttpRouteAuthorizer,
	deploymentTarget string,
) {
	// GET /me route
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/me"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integrations.GetStudentMeBySub(stack, vpc, dbParams, deploymentTarget),
		Authorizer:  Authorizer,
	})

	// GET /me/clubs route
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/me/clubs"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integrations.GetMyClubs(stack, vpc, dbParams, deploymentTarget),
		Authorizer:  Authorizer,
	})

	// GET /me/clubs route
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/me/events"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integrations.GetMyEvents(stack, vpc, dbParams, deploymentTarget),
		Authorizer:  Authorizer,
	})
}
