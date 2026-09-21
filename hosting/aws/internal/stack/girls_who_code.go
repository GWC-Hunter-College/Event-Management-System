package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type GirlsWhoCodeHostingStackProps struct {
	Props awscdk.StackProps
	// DomainName and HostedZone are optional, but must be supplied together.
	// Omit both for CloudFront-only hosting. Keep both for the live GWC site.
	DomainName string
	HostedZone awsroute53.IPublicHostedZone
}

func NewGirlsWhoCodeHostingStack(scope constructs.Construct, id string, props *GirlsWhoCodeHostingStackProps) awscdk.Stack {
	if props != nil && (props.DomainName == "") != (props.HostedZone == nil) {
		panic("GirlsWhoCodeHostingStack requires DomainName and HostedZone together, or neither")
	}
	hasCustomDomain := props != nil && props.DomainName != ""
	var stackProps awscdk.StackProps
	if props != nil {
		stackProps = props.Props
	}
	girlsWhoCodeHostingStack := awscdk.NewStack(scope, &id, &stackProps)

	girlsWhoCodeWebsiteBucket := awss3.NewBucket(girlsWhoCodeHostingStack, jsii.String("GwcWebsiteBucket"), &awss3.BucketProps{
		BucketName:        jsii.String("gwc-club-site"),
		PublicReadAccess:  jsii.Bool(false),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	awscdk.NewCfnOutput(girlsWhoCodeHostingStack, jsii.String("websiteBucketName"), &awscdk.CfnOutputProps{
		Value: girlsWhoCodeWebsiteBucket.BucketName(),
	})

	girlsWhoCodeOAC := awscloudfront.NewS3OriginAccessControl(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeOAC"), &awscloudfront.S3OriginAccessControlProps{
		Description: jsii.String("Girls Who Code at Hunter S3 origin access control"),
		Signing:     awscloudfront.Signing_SIGV4_ALWAYS(),
	})

	girlsWhoCodeStagingBehavior := &awscloudfront.BehaviorOptions{
		Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(girlsWhoCodeWebsiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{
			OriginAccessControl: girlsWhoCodeOAC,
			OriginPath:          jsii.String("/staging"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	// Keep the historical FrontendMain construct ID stable while serving the staging branch.
	girlsWhoCodeStagingDistribution := awscloudfront.NewDistribution(girlsWhoCodeHostingStack, jsii.String("FrontendMain"), &awscloudfront.DistributionProps{
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   girlsWhoCodeStagingBehavior,
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

	awscdk.NewCfnOutput(girlsWhoCodeHostingStack, jsii.String("CloudFront_Main_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Staging Branch CloudFront Info"),
		Value: jsii.String("Staging URL: https://" + *girlsWhoCodeStagingDistribution.DomainName() +
			" | ID: " + *girlsWhoCodeStagingDistribution.DistributionId()),
	})

	girlsWhoCodeProductionBehavior := &awscloudfront.BehaviorOptions{
		Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(girlsWhoCodeWebsiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{
			OriginAccessControl: girlsWhoCodeOAC,
			OriginPath:          jsii.String("/production"),
		}),
		ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
	}

	// The existing FrontendStack is deployed in us-east-1, as required by
	// CloudFront for ACM certificates. Keep that deployment environment unchanged.
	var productionDomainNames *[]*string
	var productionCertificate awscertificatemanager.Certificate
	if hasCustomDomain {
		productionDomainNames = jsii.Strings(props.DomainName, "www."+props.DomainName)
		productionCertificate = awscertificatemanager.NewCertificate(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeProductionCertificate"), &awscertificatemanager.CertificateProps{
			DomainName:              jsii.String(props.DomainName),
			SubjectAlternativeNames: jsii.Strings("www." + props.DomainName),
			Validation:              awscertificatemanager.CertificateValidation_FromDns(props.HostedZone),
		})
	}

	girlsWhoCodeProductionDistribution := awscloudfront.NewDistribution(girlsWhoCodeHostingStack, jsii.String("FrontendProduction"), &awscloudfront.DistributionProps{
		Certificate:       productionCertificate,
		DomainNames:       productionDomainNames,
		DefaultRootObject: jsii.String("index.html"),
		DefaultBehavior:   girlsWhoCodeProductionBehavior,
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

	if hasCustomDomain {
		productionAliasTarget := awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(girlsWhoCodeProductionDistribution))
		for _, record := range []struct{ id, name string }{
			{"GirlsWhoCodeProductionApex", props.DomainName},
			{"GirlsWhoCodeProductionWww", "www." + props.DomainName},
		} {
			awsroute53.NewARecord(girlsWhoCodeHostingStack, jsii.String(record.id+"A"), &awsroute53.ARecordProps{
				Zone: props.HostedZone, RecordName: jsii.String(record.name), Target: productionAliasTarget,
			})
			// The existing production distribution has IPv6 enabled (CDK's default).
			awsroute53.NewAaaaRecord(girlsWhoCodeHostingStack, jsii.String(record.id+"AAAA"), &awsroute53.AaaaRecordProps{
				Zone: props.HostedZone, RecordName: jsii.String(record.name), Target: productionAliasTarget,
			})
		}
	}

	awscdk.NewCfnOutput(girlsWhoCodeHostingStack, jsii.String("CloudFront_Production_Info"), &awscdk.CfnOutputProps{
		Description: jsii.String("Production Branch CloudFront Info"),
		Value: jsii.String("Production URL: https://" + *girlsWhoCodeProductionDistribution.DomainName() +
			" | ID: " + *girlsWhoCodeProductionDistribution.DistributionId()),
	})

	girlsWhoCodeAccount := awscdk.Stack_Of(girlsWhoCodeHostingStack).Account()
	girlsWhoCodeStagingDistributionARN := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      girlsWhoCodeAccount,
		Resource:     jsii.String("distribution"),
		ResourceName: girlsWhoCodeStagingDistribution.DistributionId(),
		Region:       jsii.String(""),
	}, girlsWhoCodeHostingStack)
	girlsWhoCodeProductionDistributionARN := awscdk.Arn_Format(&awscdk.ArnComponents{
		Service:      jsii.String("cloudfront"),
		Account:      girlsWhoCodeAccount,
		Resource:     jsii.String("distribution"),
		ResourceName: girlsWhoCodeProductionDistribution.DistributionId(),
		Region:       jsii.String(""),
	}, girlsWhoCodeHostingStack)

	girlsWhoCodeS3ObjectsStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:PutObject", "s3:DeleteObject"),
		Resources: jsii.Strings(*girlsWhoCodeWebsiteBucket.ArnForObjects(jsii.String("*"))),
	})
	girlsWhoCodeS3ListStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:ListBucket"),
		Resources: jsii.Strings(*girlsWhoCodeWebsiteBucket.BucketArn()),
	})
	girlsWhoCodeCloudFrontInvalidateStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("cloudfront:CreateInvalidation"),
		Resources: jsii.Strings(*girlsWhoCodeStagingDistributionARN, *girlsWhoCodeProductionDistributionARN),
	})

	girlsWhoCodeCIPolicy := awsiam.NewPolicy(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeCiPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String("frontend-gwc-ci-policy"),
		Statements: &[]awsiam.PolicyStatement{
			girlsWhoCodeS3ObjectsStatement,
			girlsWhoCodeS3ListStatement,
			girlsWhoCodeCloudFrontInvalidateStatement,
		},
	})

	girlsWhoCodeCIUser := awsiam.NewUser(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeCiUser"), &awsiam.UserProps{
		UserName: jsii.String("gwc-website-ci-deployer"),
	})
	girlsWhoCodeCIUser.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)
	girlsWhoCodeCIPolicy.AttachToUser(girlsWhoCodeCIUser)

	awscdk.NewCfnOutput(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeS3StagingDestination"), &awscdk.CfnOutputProps{
		Description: jsii.String("Girls Who Code staging S3 deployment destination"),
		Value:       jsii.String("s3://" + *girlsWhoCodeWebsiteBucket.BucketName() + "/staging"),
	})
	awscdk.NewCfnOutput(girlsWhoCodeHostingStack, jsii.String("GirlsWhoCodeS3ProductionDestination"), &awscdk.CfnOutputProps{
		Description: jsii.String("Girls Who Code production S3 deployment destination"),
		Value:       jsii.String("s3://" + *girlsWhoCodeWebsiteBucket.BucketName() + "/production"),
	})

	return girlsWhoCodeHostingStack
}
