package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type HunterCollegeClubsHostingStackProps struct {
	Props awscdk.StackProps
}

func NewHunterCollegeClubsHostingStack(scope constructs.Construct, id string, props *HunterCollegeClubsHostingStackProps) awscdk.Stack {
	var stackProps awscdk.StackProps
	if props != nil {
		stackProps = props.Props
	}
	hunterCollegeClubsHostingStack := awscdk.NewStack(scope, &id, &stackProps)

	hunterCollegeClubsWebsiteBucket := awss3.NewBucket(hunterCollegeClubsHostingStack, jsii.String("HunterCollegeClubEventBucket"), &awss3.BucketProps{
		BucketName:        jsii.String("hunter-college-club-event-site"),
		PublicReadAccess:  jsii.Bool(false),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	awscdk.NewCfnOutput(hunterCollegeClubsHostingStack, jsii.String("websiteBucketName"), &awscdk.CfnOutputProps{
		Value: hunterCollegeClubsWebsiteBucket.BucketName(),
	})

	hunterCollegeClubsOAC := awscloudfront.NewS3OriginAccessControl(hunterCollegeClubsHostingStack, jsii.String("HunterCollegeClubsOAC"), &awscloudfront.S3OriginAccessControlProps{
		Description: jsii.String("Hunter College Clubs / Event Manager S3 origin access control"),
		Signing:     awscloudfront.Signing_SIGV4_ALWAYS(),
	})

	hunterCollegeClubsStagingBehavior := &awscloudfront.BehaviorOptions{
		Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(hunterCollegeClubsWebsiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{
			OriginAccessControl: hunterCollegeClubsOAC,
			OriginPath:          jsii.String("/staging"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	hunterCollegeClubsStagingDistribution := awscloudfront.NewDistribution(hunterCollegeClubsHostingStack, jsii.String("FrontendStaging"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   hunterCollegeClubsStagingBehavior,
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

	awscdk.NewCfnOutput(hunterCollegeClubsHostingStack, jsii.String("CloudFront_Staging_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Staging Branch CloudFront Info"),
		Value: jsii.String("Staging URL: https://" + *hunterCollegeClubsStagingDistribution.DomainName() +
			" | ID: " + *hunterCollegeClubsStagingDistribution.DistributionId()),
	})

	hunterCollegeClubsProductionBehavior := &awscloudfront.BehaviorOptions{
		Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(hunterCollegeClubsWebsiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{
			OriginAccessControl: hunterCollegeClubsOAC,
			OriginPath:          jsii.String("/production"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	hunterCollegeClubsProductionDistribution := awscloudfront.NewDistribution(hunterCollegeClubsHostingStack, jsii.String("FrontendProduction"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   hunterCollegeClubsProductionBehavior,
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

	awscdk.NewCfnOutput(hunterCollegeClubsHostingStack, jsii.String("CloudFront_Production_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Production Branch CloudFront Info"),
		Value: jsii.String("Production URL: https://" + *hunterCollegeClubsProductionDistribution.DomainName() +
			" | ID: " + *hunterCollegeClubsProductionDistribution.DistributionId()),
	})

	hunterCollegeClubsAccount := awscdk.Stack_Of(hunterCollegeClubsHostingStack).Account()
	hunterCollegeClubsStagingDistributionARN := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      hunterCollegeClubsAccount,
		Resource:     jsii.String("distribution"),
		ResourceName: hunterCollegeClubsStagingDistribution.DistributionId(),
		Region:       jsii.String(""),
	}, hunterCollegeClubsHostingStack)
	hunterCollegeClubsProductionDistributionARN := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      hunterCollegeClubsAccount,
		Resource:     jsii.String("distribution"),
		ResourceName: hunterCollegeClubsProductionDistribution.DistributionId(),
		Region:       jsii.String(""),
	}, hunterCollegeClubsHostingStack)

	hunterCollegeClubsS3ObjectsStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:PutObject", "s3:DeleteObject"),
		Resources: jsii.Strings(*hunterCollegeClubsWebsiteBucket.ArnForObjects(jsii.String("*"))),
	})
	hunterCollegeClubsS3ListStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:ListBucket"),
		Resources: jsii.Strings(*hunterCollegeClubsWebsiteBucket.BucketArn()),
	})
	hunterCollegeClubsCloudFrontInvalidateStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("cloudfront:CreateInvalidation"),
		Resources: jsii.Strings(*hunterCollegeClubsStagingDistributionARN, *hunterCollegeClubsProductionDistributionARN),
	})

	hunterCollegeClubsCIPolicy := awsiam.NewPolicy(hunterCollegeClubsHostingStack, jsii.String("FrontendCiPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String("frontend-hcc-ci-policy"),
		Statements: &[]awsiam.PolicyStatement{
			hunterCollegeClubsS3ObjectsStatement,
			hunterCollegeClubsS3ListStatement,
			hunterCollegeClubsCloudFrontInvalidateStatement,
		},
	})

	hunterCollegeClubsCIUser := awsiam.NewUser(hunterCollegeClubsHostingStack, jsii.String("FrontendCiUser"), &awsiam.UserProps{
		UserName: jsii.String("hcc-website-ci-deployer"),
	})
	hunterCollegeClubsCIUser.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)
	hunterCollegeClubsCIPolicy.AttachToUser(hunterCollegeClubsCIUser)

	awscdk.NewCfnOutput(hunterCollegeClubsHostingStack, jsii.String("S3_Staging_Destination"), &awscdk.CfnOutputProps{
		Value: jsii.String("s3://" + *hunterCollegeClubsWebsiteBucket.BucketName() + "/staging"),
	})
	awscdk.NewCfnOutput(hunterCollegeClubsHostingStack, jsii.String("S3_Production_Destination"), &awscdk.CfnOutputProps{
		Value: jsii.String("s3://" + *hunterCollegeClubsWebsiteBucket.BucketName() + "/production"),
	})

	return hunterCollegeClubsHostingStack
}
