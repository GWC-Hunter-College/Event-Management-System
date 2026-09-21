package gateway_helpers

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
)

func GrantRdsAccessToLambda(
	lambdaFn awslambda.Function,
	dbInstance awsrds.IDatabaseInstance,
	dbSecret awssecretsmanager.ISecret,
) {
	dbInstance.GrantConnect(lambdaFn, nil)
	dbSecret.GrantRead(lambdaFn, nil)
}

func GrantS3AccessToLambda(
	lambdaFn awslambda.Function,
	bucket awss3.IBucket,
	pathPrefix string,
	grantRead bool,
	grantPut bool,
) {
	if grantRead {
		bucket.GrantRead(lambdaFn, pathPrefix)
	}

	if grantPut {
		bucket.GrantPut(lambdaFn, pathPrefix)
	}
}
