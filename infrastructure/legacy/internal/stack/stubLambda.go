package stack

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2" // core

	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"

	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StubLambdaStackProps struct {
	Props awscdk.StackProps
}

type StubLambdaStack struct {
	Stack awscdk.Stack

	PingFunction awslambda.IFunction

	StubStudentBundle StubStudentFunctions
}

type StubStudentFunctions struct {
	StubStudentsMeClubsGetEndpoint       awslambda.IFunction
	StubStudentsMeClubsEventsGetEndpoint awslambda.IFunction
	StubStudentsMeClubsEboardGetEndpoint awslambda.IFunction
}

func NewStubLambdaStack(scope constructs.Construct, id string, props *StubLambdaStackProps) *StubLambdaStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	//  =======================================
	// Api Creation
	//  =======================================
	// create HTTP API
	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("StubbedClubEventApi"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("StubbedClubEventApi"),
	})

	// The code that defines your stack goes here

	//  =======================================
	//  health
	//  =======================================
	// create health check lambda function
	pingFunction := AddStubRoute(stack, httpApi, "/health", awsapigatewayv2.HttpMethod_GET)

	//  =======================================
	//  students
	//  =======================================
	studentsMeClubsGetFunction := AddStubRoute(stack, httpApi, "/me/clubs", awsapigatewayv2.HttpMethod_GET)

	studentsMeClubsEventsGetFunction := AddStubRoute(stack, httpApi, "/me/clubs/events", awsapigatewayv2.HttpMethod_GET)

	studentsMeClubsEboardGetFunction := AddStubRoute(stack, httpApi, "/me/clubs/eboard", awsapigatewayv2.HttpMethod_GET)

	studentsBundle := StubStudentFunctions{
		StubStudentsMeClubsGetEndpoint:       studentsMeClubsGetFunction,
		StubStudentsMeClubsEventsGetEndpoint: studentsMeClubsEventsGetFunction,
		StubStudentsMeClubsEboardGetEndpoint: studentsMeClubsEboardGetFunction,
	}

	//  =======================================
	//  events
	//  =======================================
	eventsGetFunction := AddStubRoute(stack, httpApi, "/events", awsapigatewayv2.HttpMethod_GET)
	_ = eventsGetFunction

	eventsEventIdGetFunction := AddStubRoute(stack, httpApi, "/events/{eventId}", awsapigatewayv2.HttpMethod_GET)
	_ = eventsEventIdGetFunction

	eventsEventIdImagesGetFunction := AddStubRoute(stack, httpApi, "/events/{eventId}/images", awsapigatewayv2.HttpMethod_GET)
	_ = eventsEventIdImagesGetFunction

	eventsEventIdDescriptionGetFunction := AddStubRoute(stack, httpApi, "/events/{eventId}/description", awsapigatewayv2.HttpMethod_GET)
	_ = eventsEventIdDescriptionGetFunction

	eventsEventIdClubsGetFunction := AddStubRoute(stack, httpApi, "/events/{eventId}/clubs", awsapigatewayv2.HttpMethod_GET)
	_ = eventsEventIdClubsGetFunction

	//  =======================================
	//  events
	//  =======================================
	clubsGetFunction := AddStubRoute(stack, httpApi, "/clubs", awsapigatewayv2.HttpMethod_GET)
	_ = clubsGetFunction

	//  =======================================
	// 	prints
	//  =======================================
	// log HTTP API endpoint
	awscdk.NewCfnOutput(stack, jsii.String("myHttpApiEndpoint"), &awscdk.CfnOutputProps{
		Value:       httpApi.ApiEndpoint(),
		Description: jsii.String("HTTP API Endpoint"),
	})

	return &StubLambdaStack{
		Stack:        stack,
		PingFunction: pingFunction,

		StubStudentBundle: studentsBundle,
	}
}

// AddStubRoute creates a Go Lambda and wires it to the given HttpApi at `path` with `method`.
// Naming convention:
//
//	Lambda FunctionName: "Stub" + Pascal(path) + Title(method) + "Endpoint"
//	Construct ID:        "Stub " + path + " " + Title(method) + " Endpoint Function"
//	Integration ID:      "Stub" + Pascal(path) + Title(method) + "Integration"
//
// Entry path on disk:
//
//	./stub/lambda<path>/<method-lower>.go
//	e.g. path="/me/clubs/eboard", method=GET -> "./stub/lambda/me/clubs/eboard/get.go"
func AddStubRoute(scope constructs.Construct, httpApi awsapigatewayv2.HttpApi, path string, method awsapigatewayv2.HttpMethod) awslambda.IFunction {
	cleanPath := normalizePath(path)
	pascal := pathToPascal(cleanPath) // now strips {param}
	methodTitle := httpMethodTitle(method)
	methodLower := strings.ToLower(httpMethodString(method))

	lambdaName := fmt.Sprintf("Stub%s%sEndpoint", pascal, methodTitle)
	constructID := fmt.Sprintf("Stub %s %s Endpoint Function", cleanPath, methodTitle)
	integrationID := fmt.Sprintf("Stub%s%sIntegration", pascal, methodTitle)

	// NEW: use a filesystem-safe version of the path (remove braces)
	fsPath := sanitizeForFS(cleanPath) // e.g. "/events/{eventId}" -> "/events/eventId"

	entryPath := fmt.Sprintf("./stub/lambda%s/%s.go", fsPath, methodLower)
	entryPath = strings.ReplaceAll(entryPath, "//", "/")

	fn := awscdklambdagoalpha.NewGoFunction(scope, jsii.String(constructID), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String(lambdaName),
		Entry:        jsii.String(entryPath),
	})

	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:    jsii.String(cleanPath),
		Methods: &[]awsapigatewayv2.HttpMethod{method},
		Integration: awsapigatewayv2integrations.NewHttpLambdaIntegration(
			jsii.String(integrationID),
			fn,
			&awsapigatewayv2integrations.HttpLambdaIntegrationProps{},
		),
	})

	return fn
}

// ---------- small helpers ----------

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// collapse repeats
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	// keep trailing slash off except for root
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// "/me/clubs/eboard" -> "MeClubsEboard"
// "/events/{eventId}" -> "EventsEventId"
func pathToPascal(p string) string {
	trim := strings.Trim(p, "/")
	if trim == "" {
		return ""
	}
	parts := strings.Split(trim, "/")
	for i, seg := range parts {
		seg = strings.TrimSpace(seg)
		// strip {param} braces for naming
		if len(seg) >= 2 && seg[0] == '{' && seg[len(seg)-1] == '}' {
			seg = seg[1 : len(seg)-1]
		}
		parts[i] = toTitle(seg)
	}
	return strings.Join(parts, "")
}

// "/events/{eventId}" -> "/events/eventId"
// "/clubs/{clubId}/events/{eventId}" -> "/clubs/clubId/events/eventId"
func sanitizeForFS(p string) string {
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if len(seg) >= 2 && seg[0] == '{' && seg[len(seg)-1] == '}' {
			parts[i] = seg[1 : len(seg)-1] // drop braces
		}
	}
	out := strings.Join(parts, "/")
	out = strings.ReplaceAll(out, "//", "/")
	return out
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	// preserve acronyms reasonably by only uppercasing first rune
	return strings.ToUpper(s[:1]) + s[1:]
}

func httpMethodString(m awsapigatewayv2.HttpMethod) string {
	// ensure stable strings; default to underlying string
	switch m {
	case awsapigatewayv2.HttpMethod_GET:
		return "GET"
	case awsapigatewayv2.HttpMethod_POST:
		return "POST"
	case awsapigatewayv2.HttpMethod_PUT:
		return "PUT"
	case awsapigatewayv2.HttpMethod_PATCH:
		return "PATCH"
	case awsapigatewayv2.HttpMethod_DELETE:
		return "DELETE"
	default:
		return string(m)
	}
}

func httpMethodTitle(m awsapigatewayv2.HttpMethod) string {
	switch m {
	case awsapigatewayv2.HttpMethod_GET:
		return "Get"
	case awsapigatewayv2.HttpMethod_POST:
		return "Post"
	case awsapigatewayv2.HttpMethod_PUT:
		return "Put"
	case awsapigatewayv2.HttpMethod_PATCH:
		return "Patch"
	case awsapigatewayv2.HttpMethod_DELETE:
		return "Delete"
	default:
		// Title-case whatever comes through
		s := strings.ToLower(httpMethodString(m))
		return toTitle(s)
	}
}
