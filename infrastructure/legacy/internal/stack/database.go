package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/constructs-go/constructs/v10"
)

type DatabaseStackProps struct {
	Props awscdk.StackProps

	Vpc awsec2.Vpc
}

type DatabaseStack struct {
	Stack awscdk.Stack

	DbInstance awsrds.DatabaseInstance
	Proxy      awsrds.DatabaseProxy

	DbSecurityGroup     awsec2.SecurityGroup
	LambdaSecurityGroup awsec2.SecurityGroup
	ProxySecurityGroup  awsec2.SecurityGroup
}

func NewDatabaseStack(scope constructs.Construct, id string, props *DatabaseStackProps) *DatabaseStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := props.Vpc

	// proxySecurityGroup := createSecurityGroup(stack, vpc, "proxy")
	lambdaSecurityGroup := createSecurityGroup(stack, vpc, "lambda")
	dbSecurityGroup := createSecurityGroup(stack, vpc, "rds-db")

	lambdaSecurityGroup.AddEgressRule(
		dbSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(3306)),
		jsii.String("Allow connections to the database"),
		jsii.Bool(false),
	)
	// proxySecurityGroup.AddIngressRule(
	// 	lambdaSecurityGroup,
	// 	awsec2.Port_Tcp(jsii.Number(3306)),
	// 	jsii.String("Allow connections from lambda"),
	// 	jsii.Bool(false),
	// )

	// proxySecurityGroup.AddEgressRule(
	// 	dbSecurityGroup,
	// 	awsec2.Port_Tcp(jsii.Number(3306)),
	// 	jsii.String("Allow connections to the database (RDS)."),
	// 	jsii.Bool(false),
	// )
	dbSecurityGroup.AddIngressRule(
		lambdaSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(3306)),
		jsii.String("Allow connections from the lambdas"),
		jsii.Bool(false),
	)

	dbInstance := awsrds.NewDatabaseInstance(stack, jsii.String("ClubEventDb"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_8_0_37(),
		}),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		Vpc:          vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups:      &[]awsec2.ISecurityGroup{dbSecurityGroup},
		Credentials:         awsrds.Credentials_FromGeneratedSecret(jsii.String("dbadmin"), nil),
		AllocatedStorage:    jsii.Number(20),
		MaxAllocatedStorage: jsii.Number(100),
		BackupRetention:     awscdk.Duration_Days(jsii.Number(1)),
		// BackupRetention:     awscdk.Duration_Days(jsii.Number(7)), // This gave me issues with free tier
		MultiAz:            jsii.Bool(false),
		RemovalPolicy:      awscdk.RemovalPolicy_DESTROY,
		DeletionProtection: jsii.Bool(false),
	})

	// Saving costs
	// proxy := awsrds.NewDatabaseProxy(stack, jsii.String("ClubEventProxy"), &awsrds.DatabaseProxyProps{
	// 	ProxyTarget:       awsrds.ProxyTarget_FromInstance(dbInstance),
	// 	Secrets:           &[]awssecretsmanager.ISecret{dbInstance.Secret()},
	// 	Vpc:               vpc,
	// 	RequireTLS:        jsii.Bool(true),
	// 	SecurityGroups:    &[]awsec2.ISecurityGroup{proxySecurityGroup},
	// 	IdleClientTimeout: awscdk.Duration_Minutes(jsii.Number(30)),
	// })

	return &DatabaseStack{
		Stack: stack,

		DbInstance: dbInstance,
		// Proxy:      proxy,

		DbSecurityGroup:     dbSecurityGroup,
		LambdaSecurityGroup: lambdaSecurityGroup,
		// ProxySecurityGroup:  proxySecurityGroup,
	}
}
