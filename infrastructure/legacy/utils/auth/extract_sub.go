package authentication_utils

import "github.com/aws/aws-lambda-go/events"

func ExtractSubFromRequest(ctx events.APIGatewayV2HTTPRequestContext) (sub string, email string, err error) {
	claims := map[string]string{}

	if ctx.Authorizer == nil ||
		ctx.Authorizer.JWT == nil ||
		ctx.Authorizer.JWT.Claims == nil {
		err = ErrNoAuthorizer
		return
	}

	claims = ctx.Authorizer.JWT.Claims

	sub, ok := claims["sub"]
	if !ok || sub == "" {
		err = ErrNoSub
		return
	}

	email, ok = claims["email"] // email is optional
	if !ok {
		email = ""
	}

	return
}
