package stack

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"os"
	"strings"
)

type AuthenticationStackProps struct {
	Props awscdk.StackProps
}

type AuthenticationStack struct {
	Stack     awscdk.Stack
	UserPool  awscognito.IUserPool
	AppClient awscognito.IUserPoolClient
}

// .env read helpers
func toPtrs(list []string) []*string {
	out := make([]*string, 0, len(list))
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, jsii.String(s))
		}
	}
	return out
}

func load(urls string) []*string { return toPtrs(strings.Split(urls, ",")) }

func NewAuthenticationStack(scope constructs.Construct, id string, props *AuthenticationStackProps) *AuthenticationStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.Props
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	userPool := awscognito.NewUserPool(stack, jsii.String("UserPool"), &awscognito.UserPoolProps{
		SelfSignUpEnabled: jsii.Bool(true),
		SignInAliases: &awscognito.SignInAliases{
			Email: jsii.Bool(true),
		},
		StandardAttributes: &awscognito.StandardAttributes{
			Email: &awscognito.StandardAttribute{
				Required: jsii.Bool(true),
				Mutable:  jsii.Bool(true),
			},
		},
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:            jsii.Number(7),
			RequireLowercase:     jsii.Bool(false),
			RequireUppercase:     jsii.Bool(false),
			RequireDigits:        jsii.Bool(false),
			RequireSymbols:       jsii.Bool(false),
			TempPasswordValidity: awscdk.Duration_Days(jsii.Number(7)),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // <- forces deletion
	})

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolId"), &awscdk.CfnOutputProps{
		Value: userPool.UserPoolId(),
	})

	//

	domain := userPool.AddDomain(jsii.String("CognitoDomain"), &awscognito.UserPoolDomainOptions{
		CognitoDomain: &awscognito.CognitoDomainOptions{
			DomainPrefix: jsii.String("event-manager-authz"),
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("HostedUIDomainBaseUrl"), &awscdk.CfnOutputProps{
		// e.g., https://gwc-auth-yourprefix.auth.us-east-1.amazoncognito.com
		Value: domain.BaseUrl(&awscognito.BaseUrlOptions{}),
	})

	googleProvider := awscognito.NewUserPoolIdentityProviderGoogle(stack, jsii.String("GoogleIdP"), &awscognito.UserPoolIdentityProviderGoogleProps{
		UserPool:     userPool,
		ClientId:     jsii.String(os.Getenv("GOOGLE_CLIENT_ID")),
		ClientSecret: jsii.String(os.Getenv("GOOGLE_CLIENT_SECRET")),
		Scopes: &[]*string{
			jsii.String("openid"),
			jsii.String("email"),
			jsii.String("profile"),
		},
		AttributeMapping: &awscognito.AttributeMapping{
			Email:          awscognito.ProviderAttribute_GOOGLE_EMAIL(),
			GivenName:      awscognito.ProviderAttribute_GOOGLE_GIVEN_NAME(),
			FamilyName:     awscognito.ProviderAttribute_GOOGLE_FAMILY_NAME(),
			ProfilePicture: awscognito.ProviderAttribute_GOOGLE_PICTURE(),
		},
	})

	//

	callbackUrls := load(os.Getenv("CALLBACK_URLS"))
	logoutUrls := load(os.Getenv("LOGOUT_URLS"))

	webClient := userPool.AddClient(jsii.String("WebClient"), &awscognito.UserPoolClientOptions{
		GenerateSecret: jsii.Bool(false), // public client (browser)
		AuthFlows: &awscognito.AuthFlow{
			UserPassword: jsii.Bool(true),
		},
		PreventUserExistenceErrors: jsii.Bool(true),
		OAuth: &awscognito.OAuthSettings{
			Flows: &awscognito.OAuthFlows{
				AuthorizationCodeGrant: jsii.Bool(true),
			},
			Scopes: &[]awscognito.OAuthScope{
				awscognito.OAuthScope_OPENID(),
				awscognito.OAuthScope_EMAIL(),
				awscognito.OAuthScope_PROFILE(),
			},
			CallbackUrls: &callbackUrls,
			LogoutUrls:   &logoutUrls,
		},
		SupportedIdentityProviders: &[]awscognito.UserPoolClientIdentityProvider{
			awscognito.UserPoolClientIdentityProvider_COGNITO(),
			// google support
			awscognito.UserPoolClientIdentityProvider_GOOGLE(),
		},
		// Recommended for browsers
		// DisableOAuthScopesRequired: jsii.Bool(false),
		EnableTokenRevocation: jsii.Bool(true),
		AccessTokenValidity:   awscdk.Duration_Hours(jsii.Number(1)),
		IdTokenValidity:       awscdk.Duration_Hours(jsii.Number(1)),
		RefreshTokenValidity:  awscdk.Duration_Days(jsii.Number(30)),
	})
	webClient.Node().AddDependency(googleProvider)

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolClientId"), &awscdk.CfnOutputProps{
		Value: webClient.UserPoolClientId(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("CallbackUrls"), &awscdk.CfnOutputProps{
		Value: jsii.String(os.Getenv("CALLBACK_URLS")),
	})
	awscdk.NewCfnOutput(stack, jsii.String("LogoutUrls"), &awscdk.CfnOutputProps{
		Value: jsii.String(os.Getenv("LOGOUT_URLS")),
	})

	// //

	// awscdk.NewCfnOutput(stack, jsii.String("AuthorizeUrlTemplate"), &awscdk.CfnOutputProps{
	// 	// Helpful to test the redirect immediately
	// 	Value: jsii.String("{{HostedUIDomainBaseUrl}}/oauth2/authorize?client_id={{UserPoolClientId}}&response_type=code&scope=openid+email+profile&redirect_uri=http://localhost:5173/callback"),
	// })

	// redirectUri := "http://localhost:5173/callback"
	// awscdk.NewCfnOutput(stack, jsii.String("AuthorizeUrl"), &awscdk.CfnOutputProps{
	// 	Value: jsii.String(
	// 		*domain.BaseUrl(&awscognito.BaseUrlOptions{}) +
	// 			"/oauth2/authorize" +
	// 			"?client_id=" + *webClient.UserPoolClientId() +
	// 			"&response_type=code" +
	// 			"&scope=openid+email+profile" +
	// 			"&redirect_uri=" + url.QueryEscape(redirectUri),
	// 	),
	// })

	return &AuthenticationStack{
		Stack:     stack,
		UserPool:  userPool,
		AppClient: webClient,
	}
}
