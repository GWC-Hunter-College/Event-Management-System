package main

import (
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/joho/godotenv"

	"github.com/aws/jsii-runtime-go"

	stack "cdk-infrastructure/internal/stack"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, relying on system env vars")
	}

	stack.NewFrontendStack(app, "FrontendStack", &stack.FrontendStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the GWC website deployment"),
		},
	})

	stack.NewFrontendHccStack(app, "FrontendHccStack", &stack.FrontendHccStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the EMS website deployment"),
		},
	})

	network := stack.NewNetworkStack(app, "NetworkStack", &stack.NetworkStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for all network infrastructure constructs"),
		},
	})

	database := stack.NewDatabaseStack(app, "DatabaseStack", &stack.DatabaseStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for database constructs"),
		},
		Vpc: network.Vpc,
	})

	image := stack.NewImageStack(app, "ImageStack", &stack.ImageStackProps{
		Props: awscdk.StackProps{
			Description: jsii.String("Stack for all images related to the events system"),
			Env:         env(),
		},
	})

	authentication := stack.NewAuthenticationStack(app, "AuthenticationStack", &stack.AuthenticationStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for Cognito User Pool and App Client"),
		},
	})

	authorization := stack.NewAuthorizationStack(app, "AuthorizationStack", &stack.AuthorizationStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for Cognito Authorizer for API Gateway"),
		},

		Vpc:                               network.Vpc,
		LambdaSecretsManagerSecurityGroup: network.LambdaSecretsManagerSecurityGroup,
		DbInstance:                        database.DbInstance,
		LambdaSecurityGroup:               database.LambdaSecurityGroup,

		UserPool:  authentication.UserPool,
		AppClient: authentication.AppClient,
	})

	stack.NewProdApiStack(app, "ProdApiStack", &stack.ProdApiStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the production API Gateway and its routes"),
		},

		Vpc:                               network.Vpc,
		LambdaSecretsManagerSecurityGroup: network.LambdaSecretsManagerSecurityGroup,
		LambdaSecurityGroup:               database.LambdaSecurityGroup,
		DbInstance:                        database.DbInstance,

		Bucket: image.Bucket,

		Authorizer: authorization.Authorizer,
	})

	stack.NewDevApiStack(app, "DevApiStack", &stack.DevApiStackProps{
		Props: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the development API Gateway and its routes"),
		},

		Vpc:                               network.Vpc,
		LambdaSecretsManagerSecurityGroup: network.LambdaSecretsManagerSecurityGroup,
		LambdaSecurityGroup:               database.LambdaSecurityGroup,
		DbInstance:                        database.DbInstance,

		Bucket: image.Bucket,

		UserPool:  authentication.UserPool,
		AppClient: authentication.AppClient,
	})

	stack.NewBastionStack(app, "BastionStack", &stack.BastionStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the RDS root access bastion EC2 instance."),
		},

		Vpc:             network.Vpc,
		DbSecurityGroup: database.DbSecurityGroup,
	})

	// stack.NewStubLambdaStack(app, "StubLambdaStack", &stack.StubLambdaStackProps{
	// 	Props: awscdk.StackProps{
	// 		Env: env(),
	// 	},
	// })

	stack.NewDatabaseInitStack(app, "DatabaseInitStack", &stack.Props{
		StackProps: awscdk.StackProps{
			Env:         env(),
			Description: jsii.String("Stack for the database table initialization lambdas. Should destroy if you see it since the lambda has already run."),
		},

		Vpc:                               network.Vpc,
		LambdaSecretsManagerSecurityGroup: network.LambdaSecretsManagerSecurityGroup,
		LambdaSecurityGroup:               database.LambdaSecurityGroup,
		DbInstance:                        database.DbInstance,
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
