package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
)

type FrontendHccStackProps struct {
	Props awscdk.StackProps
}

func NewFrontendHccStack(scope constructs.Construct, id string, props *FrontendHccStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	websiteBucket := awss3.NewBucket(stack,
		jsii.String("HunterCollegeClubEventBucket"), // logical ID
		&awss3.BucketProps{
			BucketName:        jsii.String("hunter-college-club-event-site"),
			PublicReadAccess:  jsii.Bool(false),
			RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
			AutoDeleteObjects: jsii.Bool(true),
		})

	// Output S3 bucket name
	awscdk.NewCfnOutput(stack, jsii.String("websiteBucketName"), &awscdk.CfnOutputProps{
		Value: websiteBucket.BucketName(),
	})

	// authorization for the s3 bucket through oai
	cloudfrontOAI := awscloudfront.NewOriginAccessIdentity(stack, jsii.String("FrontendOAI"), &awscloudfront.OriginAccessIdentityProps{})

	websiteBucket.GrantRead(cloudfrontOAI, nil)

	// first the main (can also be called staging)
	// cloudfront configeration
	cloudfrontStagingBehavior := &awscloudfront.BehaviorOptions{
		// Sets the S3 Bucket as the origin and tells CloudFront to use the created OAI to access it
		Origin: awscloudfrontorigins.NewS3Origin(websiteBucket, &awscloudfrontorigins.S3OriginProps{
			OriginAccessIdentity: cloudfrontOAI,
			OriginPath:           jsii.String("/staging"),
			// OriginId:             jsii.String("CloudFrontS3Access"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	var frontendStaging awscloudfront.Distribution

	frontendStaging = awscloudfront.NewDistribution(stack, jsii.String("FrontendStaging"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   cloudfrontStagingBehavior,
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(0)),
			},
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(0)),
			},
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("CloudFront_Staging_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Staging Branch CloudFront Info"),
		Value: jsii.String("Staging URL: https://" + *frontendStaging.DomainName() +
			" | ID: " + *frontendStaging.DistributionId()),
	})

	// first the main (can also be called staging)
	// cloudfront configeration
	cloudfrontProductionBehavior := &awscloudfront.BehaviorOptions{
		// Sets the S3 Bucket as the origin and tells CloudFront to use the created OAI to access it
		Origin: awscloudfrontorigins.NewS3Origin(websiteBucket, &awscloudfrontorigins.S3OriginProps{
			OriginAccessIdentity: cloudfrontOAI,
			OriginPath:           jsii.String("/production"),
			// OriginId:             jsii.String("CloudFrontS3Access"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	var frontendProduction awscloudfront.Distribution

	frontendProduction = awscloudfront.NewDistribution(stack, jsii.String("FrontendProduction"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   cloudfrontProductionBehavior,
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(0)),
			},
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(0)),
			},
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("CloudFront_Production_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Production Branch CloudFront Info"),
		Value: jsii.String("Production URL: https://" + *frontendProduction.DomainName() +
			" | ID: " + *frontendProduction.DistributionId()),
	})

	// create IAM user for the account as well as a policy to attach to the IAM
	// to allow for the IAM to have acces to s3 and cloudfront for ci/cd

	// iam user is account specific so must grab account
	account := awscdk.Stack_Of(stack).Account()

	// grab the arns of the distributions to attach to policy
	// stagingDistArn := awscdk.Arn_Format(&awscdk.ArnComponents{
	// 	Service:      jsii.String("cloudfront"),
	// 	Account:      account,
	// 	Resource:     jsii.String("distribution"),
	// 	ResourceName: frontendStaging.DistributionId(),
	// }, stack)
	// prodDistArn := awscdk.Arn_Format(&awscdk.ArnComponents{
	// 	Service:      jsii.String("cloudfront"),
	// 	Account:      account,
	// 	Resource:     jsii.String("distribution"),
	// 	ResourceName: frontendProduction.DistributionId(),
	// }, stack)
	stagingDistArn := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      account,
		Resource:     jsii.String("distribution"),
		ResourceName: frontendStaging.DistributionId(),
		Region:       jsii.String(""),
	}, stack)
	prodDistArn := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      account,
		Resource:     jsii.String("distribution"),
		ResourceName: frontendProduction.DistributionId(),
		Region:       jsii.String(""),
	}, stack)

	// create polices for s3 access and cloudfront invalidations

	// S3 policy statements (scoped to this bucket)
	s3ObjectsStmt := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:PutObject", "s3:DeleteObject"),
		Resources: jsii.Strings(*websiteBucket.ArnForObjects(jsii.String("*"))),
	})
	s3ListStmt := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:ListBucket"),
		Resources: jsii.Strings(*websiteBucket.BucketArn()),
	})

	// CloudFront invalidation for both dists
	cfInvalidateStmt := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("cloudfront:CreateInvalidation"),
		Resources: jsii.Strings(*stagingDistArn, *prodDistArn),
	})

	ciPolicy := awsiam.NewPolicy(stack, jsii.String("FrontendCiPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String("frontend-hcc-ci-policy"),
		Statements: &[]awsiam.PolicyStatement{s3ObjectsStmt, s3ListStmt, cfInvalidateStmt},
	})

	// create user and attach policy
	ciUser := awsiam.NewUser(stack, jsii.String("FrontendCiUser"), &awsiam.UserProps{
		UserName: jsii.String("hcc-website-ci-deployer"),
	})
	ciPolicy.AttachToUser(ciUser)

	//
	awscdk.NewCfnOutput(stack, jsii.String("S3_Staging_Destination"), &awscdk.CfnOutputProps{
		Value: jsii.String("s3://" + *websiteBucket.BucketName() + "/staging"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("S3_Production_Destination"), &awscdk.CfnOutputProps{
		Value: jsii.String("s3://" + *websiteBucket.BucketName() + "/production"),
	})

	return stack
}
