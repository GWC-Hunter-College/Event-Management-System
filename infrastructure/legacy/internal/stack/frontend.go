package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2" // core
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
)

type FrontendStackProps struct {
	Props awscdk.StackProps
}

func NewFrontendStack(scope constructs.Construct, id string, props *FrontendStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	websiteBucket := awss3.NewBucket(stack,
		jsii.String("GwcWebsiteBucket"), // logical ID
		&awss3.BucketProps{
			BucketName:        jsii.String("gwc-club-site"),
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
	cloudfrontMainBehavior := &awscloudfront.BehaviorOptions{
		// Sets the S3 Bucket as the origin and tells CloudFront to use the created OAI to access it
		Origin: awscloudfrontorigins.NewS3Origin(websiteBucket, &awscloudfrontorigins.S3OriginProps{
			OriginAccessIdentity: cloudfrontOAI,
			OriginPath:           jsii.String("/main"),
			// OriginId:             jsii.String("CloudFrontS3Access"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	var frontendMain awscloudfront.Distribution

	frontendMain = awscloudfront.NewDistribution(stack, jsii.String("FrontendMain"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   cloudfrontMainBehavior,
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

	awscdk.NewCfnOutput(stack, jsii.String("CloudFront_Main_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Main Branch CloudFront Info"),
		Value: jsii.String("Main URL: https://" + *frontendMain.DomainName() +
			" | ID: " + *frontendMain.DistributionId()),
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

	return stack
}
