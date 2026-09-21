package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type GirlsWhoCodeDomainStackProps struct {
	Props      awscdk.StackProps
	DomainName string
}

type GirlsWhoCodeDomainStack struct {
	awscdk.Stack
	HostedZone awsroute53.IPublicHostedZone
}

func NewGirlsWhoCodeDomainStack(scope constructs.Construct, id string, props *GirlsWhoCodeDomainStackProps) *GirlsWhoCodeDomainStack {
	if props == nil || props.DomainName == "" {
		panic("GirlsWhoCodeDomainStack requires DomainName")
	}
	domainStack := awscdk.NewStack(scope, &id, &props.Props)
	hostedZone := awsroute53.NewPublicHostedZone(domainStack, jsii.String("GirlsWhoCodeHostedZone"), &awsroute53.PublicHostedZoneProps{
		ZoneName: jsii.String(props.DomainName),
	})
	// Keep the authoritative zone if the stack is accidentally removed.
	hostedZone.ApplyRemovalPolicy(awscdk.RemovalPolicy_RETAIN)

	awscdk.NewCfnOutput(domainStack, jsii.String("HostedZoneId"), &awscdk.CfnOutputProps{
		Value: hostedZone.HostedZoneId(),
	})
	awscdk.NewCfnOutput(domainStack, jsii.String("NameServers"), &awscdk.CfnOutputProps{
		Description: jsii.String("Enter these four authoritative nameservers in Namecheap Custom DNS"),
		Value:       awscdk.Fn_Join(jsii.String(", "), hostedZone.HostedZoneNameServers()),
	})

	return &GirlsWhoCodeDomainStack{Stack: domainStack, HostedZone: hostedZone}
}
