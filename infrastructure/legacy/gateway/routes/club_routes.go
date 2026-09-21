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
//	GET    /clubs
//	GET    /clubs/{clubId}
//	GET    /clubs/{clubId}/events
//	POST   /clubs/{clubId}/events
//	POST   /clubs/{clubId}/members/me
//	DELETE /clubs/{clubId}/members/me
func ClubRoutes(
	httpApi awsapigatewayv2.HttpApi,
	stack awscdk.Stack,
	vpc awsec2.Vpc,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	authorizer awsapigatewayv2.IHttpRouteAuthorizer,
	deploymentTarget string,
) {
	// GET /clubs
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
		Integration: integrations.GetClubsByVerification(stack, vpc, dbParams, deploymentTarget),
	})

	// GET /clubs/{clubId}
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
		Integration: integrations.GetClubById(stack, vpc, dbParams, deploymentTarget),
	})

	// GET /clubs/{clubId}/events
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}/events"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
		Integration: integrations.GetClubEventsByPeriod(stack, vpc, dbParams, deploymentTarget),
	})

	// POST /clubs/{clubId}/events
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}/events"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_POST,
		},
		Integration: integrations.PostNewClubEvent(stack, vpc, dbParams, deploymentTarget),
		Authorizer:  authorizer,
	})

	// POST /clubs/{clubId}/members/me  (join club as current student)
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}/members/me"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_POST,
		},
		Integration: integrations.PostJoinClubMemberMe(stack, vpc, dbParams),
		Authorizer:  authorizer,
	})

	// DELETE /clubs/{clubId}/members/me  (leave club as current student)
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}/members/me"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_DELETE,
		},
		Integration: integrations.DeleteClubMemberMe(stack, vpc, dbParams),
	})

	// POST /clubs
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_POST,
		},
		Integration: integrations.PostNewClub(stack, vpc, dbParams, deploymentTarget),
		Authorizer:  authorizer,
	})
}
