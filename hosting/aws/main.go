package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"

	"github.com/GWC-Hunter-College/cdk-infrastructure/hosting/aws/internal/stack"
)

// Keep the historical CDK stack IDs until existing CloudFormation ownership is
// reviewed. The Go symbols remain explicit about the application each stack hosts.
const (
	girlsWhoCodeDomainName           = "girlswhocodehunter.org"
	girlsWhoCodeDomainStackID        = "GirlsWhoCodeDomainStack"
	girlsWhoCodeHostingStackID       = "FrontendStack"
	hunterCollegeClubsHostingStackID = "FrontendHccStack"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	domainStack := stack.NewGirlsWhoCodeDomainStack(app, girlsWhoCodeDomainStackID, &stack.GirlsWhoCodeDomainStackProps{
		Props: awscdk.StackProps{
			Env:         environment(),
			Description: jsii.String("Girls Who Code at Hunter authoritative DNS zone"),
		},
		DomainName: girlsWhoCodeDomainName,
	})

	stack.NewGirlsWhoCodeHostingStack(app, girlsWhoCodeHostingStackID, &stack.GirlsWhoCodeHostingStackProps{
		Props: awscdk.StackProps{
			Env:         environment(),
			Description: jsii.String("Girls Who Code at Hunter website hosting"),
		},
		DomainName: girlsWhoCodeDomainName,
		HostedZone: domainStack.HostedZone,
	})

	stack.NewHunterCollegeClubsHostingStack(app, hunterCollegeClubsHostingStackID, &stack.HunterCollegeClubsHostingStackProps{
		Props: awscdk.StackProps{
			Env:         environment(),
			Description: jsii.String("Hunter College Clubs / Event Manager website hosting"),
		},
	})

	app.Synth(nil)
}

// environment leaves the stacks environment-agnostic, matching the reference
// hosting application. Account and region are resolved when the stacks deploy.
func environment() *awscdk.Environment {
	return nil
}
