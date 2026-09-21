package integrations

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"
)

type DatabaseTestIntegrationProps struct {
	Vpc                               awsec2.Vpc
	LambdaSecretsManagerSecurityGroup awsec2.SecurityGroup
	LambdaSecurityGroup               awsec2.SecurityGroup
	DbInstance                        awsrds.DatabaseInstance
	ProxyEndpoint                     *string
}

func DatabaseTestIntegration(stack awscdk.Stack, props DatabaseTestIntegrationProps) awsapigatewayv2integrations.HttpLambdaIntegration {
	dbTestFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("DBTestFunction"), &awscdklambdagoalpha.GoFunctionProps{
		Entry:      jsii.String("lambda/api/database/test/main.go"), // path to folder with main.go
		MemorySize: jsii.Number(256),
		Timeout:    awscdk.Duration_Seconds(jsii.Number(10)),
		Environment: &map[string]*string{
			"DB_SECRET_ARN": props.DbInstance.Secret().SecretArn(),
			"DB_HOST":       props.ProxyEndpoint,
		},
		Vpc: props.Vpc,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			props.LambdaSecretsManagerSecurityGroup,
			props.LambdaSecurityGroup,
		},
		AllowPublicSubnet: jsii.Bool(true),
	})

	props.DbInstance.Secret().GrantRead(dbTestFunction, nil)

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("DBTestIntegration"),
		dbTestFunction,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
	)

	return integration
}
