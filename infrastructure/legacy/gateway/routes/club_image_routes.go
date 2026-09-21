package gateway_routes

import (
	gateway_helpers "cdk-infrastructure/gateway/helpers"
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
// GET  /clubs/{clubId}/thumbnails
// POST /clubs/{clubId}/thumbnails
func ClubImageRoutes(
	httpApi awsapigatewayv2.HttpApi,
	stack awscdk.Stack,
	vpc awsec2.Vpc,
	s3Params gateway_parameters.S3PermissionsParameters,
	dbParams gateway_parameters.DatabaseConnectionParameters,
	deploymentTarget string,
) {
	// proxySg := dbParams.RdsProxySG
	// dbInstance := dbParams.DbInstance
	// dbSecret := dbParams.Secret
	bucket := s3Params.Bucket

	// GET /clubs/{clubId}/thumbnails
	// getThumbnailsInt, getThumbnailsFn := integrations.GetClubThumbnailsIntegration(stack, s3Params)
	// gateway_helpers.GrantRdsProxyAccessToLambda(postThumbnailsFn, proxySg, dbInstance, dbSecret)
	// httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
	// 	Path: jsii.String("/clubs/{clubId}/thumbnails"),
	// 	Methods: &[]awsapigatewayv2.HttpMethod{
	// 		awsapigatewayv2.HttpMethod_GET,
	// 	},
	// 	Integration: integrations.GetClubThumbnailsIntegration(stack, clubImagesBucket),
	// })

	// POST /clubs/{clubId}/thumbnails
	postThumbnailsInt, postThumbnailsFunc := integrations.PostClubThumbnailsIntegration(stack, vpc, s3Params, deploymentTarget)
	gateway_helpers.GrantS3AccessToLambda(postThumbnailsFunc, bucket, "clubs/*/thumbnails/*", false, true)

	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path: jsii.String("/clubs/{clubId}/thumbnails"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_POST,
		},
		Integration: postThumbnailsInt,
	})
}
