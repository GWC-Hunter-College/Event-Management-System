package gateway_parameters

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
)

type DatabaseConnectionParameters struct {
	Secret                   awssecretsmanager.ISecret
	LambdaSG                 awsec2.ISecurityGroup
	LambdaToSecretsManagerSG awsec2.ISecurityGroup
	DbHost                   string
	DbInstance               awsrds.DatabaseInstance
	DbName                   string
}

type S3PermissionsParameters struct {
	Bucket awss3.IBucket
}
