package stack

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestMain(m *testing.M) {
	code := m.Run()
	jsii.Close()
	os.Exit(code)
}

func TestGirlsWhoCodeProductionDomain(t *testing.T) {
	app := awscdk.NewApp(&awscdk.AppProps{Outdir: jsii.String(t.TempDir())})
	domain := NewGirlsWhoCodeDomainStack(app, "GirlsWhoCodeDomainStack", &GirlsWhoCodeDomainStackProps{
		DomainName: "example.org",
	})
	frontend := NewGirlsWhoCodeHostingStack(app, "FrontendStack", &GirlsWhoCodeHostingStackProps{
		DomainName: "example.org", HostedZone: domain.HostedZone,
	})
	zoneTemplate := assertions.Template_FromStack(domain.Stack, nil)
	template := assertions.Template_FromStack(frontend, nil)
	zoneTemplate.ResourceCountIs(jsii.String("AWS::Route53::HostedZone"), jsii.Number(1))
	zoneTemplate.HasResource(jsii.String("AWS::Route53::HostedZone"), map[string]interface{}{
		"Properties":          map[string]interface{}{"Name": "example.org.", "VPCs": assertions.Match_Absent()},
		"DeletionPolicy":      "Retain",
		"UpdateReplacePolicy": "Retain",
	})
	if *zoneTemplate.GetResourceId(jsii.String("AWS::Route53::HostedZone"), nil) != "GirlsWhoCodeHostedZoneDBA9EB09" {
		t.Fatal("existing hosted zone logical ID changed")
	}
	zoneTemplate.HasOutput(jsii.String("HostedZoneId"), map[string]interface{}{"Value": assertions.Match_AnyValue()})
	zoneTemplate.HasOutput(jsii.String("NameServers"), map[string]interface{}{
		"Value": map[string]interface{}{"Fn::Join": assertions.Match_AnyValue()},
	})
	if len(*domain.Dependencies()) != 0 {
		t.Fatal("domain stack must not depend on the frontend stack")
	}
	if dependencies := *frontend.Dependencies(); len(dependencies) != 1 || dependencies[0] != domain.Stack {
		t.Fatal("frontend must depend only on the domain stack")
	}

	zoneImport := map[string]interface{}{"Fn::ImportValue": assertions.Match_AnyValue()}
	template.ResourceCountIs(jsii.String("AWS::CertificateManager::Certificate"), jsii.Number(1))
	if *template.GetResourceId(jsii.String("AWS::CertificateManager::Certificate"), nil) != "GirlsWhoCodeProductionCertificate70C3275C" {
		t.Fatal("existing certificate logical ID changed")
	}
	template.HasResourceProperties(jsii.String("AWS::CertificateManager::Certificate"), map[string]interface{}{
		"DomainName": "example.org", "SubjectAlternativeNames": []string{"www.example.org"},
		"ValidationMethod": "DNS",
		"DomainValidationOptions": []interface{}{
			map[string]interface{}{"DomainName": "example.org", "HostedZoneId": zoneImport},
			map[string]interface{}{"DomainName": "www.example.org", "HostedZoneId": zoneImport},
		},
	})
	template.ResourceCountIs(jsii.String("AWS::CloudFront::Distribution"), jsii.Number(2))
	productionID := template.GetResourceId(jsii.String("AWS::CloudFront::Distribution"), map[string]interface{}{
		"Properties": map[string]interface{}{"DistributionConfig": map[string]interface{}{
			"Aliases": []string{"example.org", "www.example.org"}, "IPV6Enabled": true,
			"ViewerCertificate": map[string]interface{}{"AcmCertificateArn": assertions.Match_AnyValue()},
		}},
	})
	if *productionID != "FrontendProduction57D7F36D" {
		t.Fatal("existing production distribution logical ID changed")
	}
	stagingID := template.GetResourceId(jsii.String("AWS::CloudFront::Distribution"), map[string]interface{}{
		"Properties": map[string]interface{}{"DistributionConfig": map[string]interface{}{
			"Aliases": assertions.Match_Absent(), "ViewerCertificate": assertions.Match_Absent(),
		}},
	})
	if *stagingID != "FrontendMain4FAF8302" {
		t.Fatal("existing staging distribution logical ID changed")
	}
	template.ResourceCountIs(jsii.String("AWS::Route53::RecordSet"), jsii.Number(4))
	for _, name := range []string{"example.org.", "www.example.org."} {
		for _, recordType := range []string{"A", "AAAA"} {
			template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
				"Name": name, "Type": recordType, "HostedZoneId": zoneImport,
				"AliasTarget": map[string]interface{}{
					"DNSName": map[string]interface{}{"Fn::GetAtt": []string{*productionID, "DomainName"}},
				},
			})
		}
	}
}

func TestDomainlessHosting(t *testing.T) {
	for _, site := range []string{"gwc-nil-props", "gwc-stack-props", "hunter-clubs"} {
		t.Run(site, func(t *testing.T) {
			// Each app contains only a hosting stack: there is no DomainStack to import.
			app := awscdk.NewApp(&awscdk.AppProps{Outdir: jsii.String(t.TempDir())})
			var hosting awscdk.Stack
			switch site {
			case "gwc-nil-props":
				hosting = NewGirlsWhoCodeHostingStack(app, "FrontendStack", nil)
			case "gwc-stack-props":
				hosting = NewGirlsWhoCodeHostingStack(app, "FrontendStack", &GirlsWhoCodeHostingStackProps{
					Props: awscdk.StackProps{Description: jsii.String("Domainless hosting")},
				})
			case "hunter-clubs":
				hosting = NewHunterCollegeClubsHostingStack(app, "FrontendHccStack", nil)
			}
			template := assertions.Template_FromStack(hosting, nil)
			template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(1))
			template.ResourceCountIs(jsii.String("AWS::CloudFront::Distribution"), jsii.Number(2))
			template.ResourceCountIs(jsii.String("AWS::CloudFront::OriginAccessControl"), jsii.Number(1))
			template.ResourceCountIs(jsii.String("AWS::IAM::User"), jsii.Number(1))
			template.ResourceCountIs(jsii.String("AWS::IAM::Policy"), jsii.Number(1))
			for _, resourceType := range []string{"AWS::Route53::HostedZone", "AWS::Route53::RecordSet", "AWS::CertificateManager::Certificate"} {
				template.ResourceCountIs(jsii.String(resourceType), jsii.Number(0))
			}
			template.AllResourcesProperties(jsii.String("AWS::CloudFront::Distribution"), map[string]interface{}{
				"DistributionConfig": map[string]interface{}{
					"Aliases": assertions.Match_Absent(), "ViewerCertificate": assertions.Match_Absent(),
				},
			})
			template.HasOutput(jsii.String("CloudFrontProductionInfo"), map[string]interface{}{
				"Value": map[string]interface{}{"Fn::Join": assertions.Match_AnyValue()},
			})
			if len(*hosting.Dependencies()) != 0 {
				t.Fatal("domainless hosting must have no stack dependencies")
			}
			encoded, err := json.Marshal(template.ToJSON())
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "Fn::ImportValue") {
				t.Fatal("domainless hosting must have no CloudFormation imports")
			}
		})
	}
}

func TestIncompleteCustomDomainRejected(t *testing.T) {
	app := awscdk.NewApp(&awscdk.AppProps{Outdir: jsii.String(t.TempDir())})
	domain := NewGirlsWhoCodeDomainStack(app, "Domain", &GirlsWhoCodeDomainStackProps{DomainName: "example.org"})
	for _, props := range []*GirlsWhoCodeHostingStackProps{
		{DomainName: "example.org"},
		{HostedZone: domain.HostedZone},
	} {
		func() {
			defer func() {
				if got := recover(); got != "GirlsWhoCodeHostingStack requires DomainName and HostedZone together, or neither" {
					t.Errorf("expected incomplete domain validation, got %v", got)
				}
			}()
			NewGirlsWhoCodeHostingStack(app, "InvalidFrontend", props)
		}()
	}
}
