package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ImageStackProps struct {
	Props awscdk.StackProps
}

type ImageStack struct {
	Stack  awscdk.Stack
	Bucket awss3.Bucket
}

func NewImageStack(scope constructs.Construct, id string, props *ImageStackProps) *ImageStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	imagesBucket := awss3.NewBucket(stack, jsii.String("hc-images-bucket"), &awss3.BucketProps{
		BucketName: jsii.String("hunter-event-sys-uploaded-images"),
		Cors: &[]*awss3.CorsRule{
			{
				AllowedMethods: &[]awss3.HttpMethods{
					awss3.HttpMethods_GET,
					awss3.HttpMethods_PUT,
					awss3.HttpMethods_POST,
				},
				AllowedOrigins: &[]*string{
					jsii.String("*"), // allowing from all origins atm, should be locked down later
				},
				AllowedHeaders: &[]*string{
					jsii.String("*"),
				},
			},
		},
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EnforceSSL:        jsii.Bool(true),
	})

	return &ImageStack{
		Stack:  stack,
		Bucket: imagesBucket,
	}
}
