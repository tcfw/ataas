package utils

import (
	"context"
	"fmt"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"pm.tcfw.com.au/source/trader/api/pb/passport"
)

//ValidateAuthToken validates a token using the passport service
func ValidateAuthToken(ctx context.Context, token string) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, fmt.Errorf("failed to parse metadata")
	}

	auth := md.Get("authorization")
	if len(auth) == 0 {
		return false, fmt.Errorf("no authorization sent")
	}

	passportEndpoint, envExists := os.LookupEnv("PASSPORT_HOST")
	if !envExists {
		passportEndpoint = viper.GetString("grpc.addr")
	}

	conn, err := grpc.DialContext(ctx, passportEndpoint, grpc.WithInsecure())
	if err != nil {
		return false, err
	}

	passportSvc := passport.NewPassportSeviceClient(conn)

	tokenResponse, err := passportSvc.VerifyToken(ctx, &passport.VerifyTokenRequest{Token: auth[0]})
	if err != nil {
		return false, err
	}
	if !tokenResponse.Valid || tokenResponse.Revoked {
		return false, fmt.Errorf("invalid auth token provided")
	}

	return true, nil
}

//GetAuthToken extracts the token from the given context
func GetAuthToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to parse metadata")
	}

	auth := md.Get("grpcgateway-authorization")

	if len(auth) == 0 {
		auth = md.Get("authorization")
		if len(auth) == 0 {
			return "", fmt.Errorf("no authorization sent in context")
		}
	}

	return auth[0], nil
}

//ValidateAuthClaims validates a token in the Authorization header and returns claims map
func ValidateAuthClaims(ctx context.Context) (jwt.MapClaims, error) {
	token, err := GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	valid, err := ValidateAuthToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid token provided")
	}

	return TokenClaims(token)
}

//TokenClaimsFromContext fetches token claims via context
func TokenClaimsFromContext(ctx context.Context) (map[string]interface{}, error) {
	token, err := GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	return TokenClaims(token)
}

//TokenClaims parses a JWT token and returns it's body claims
func TokenClaims(token string) (map[string]interface{}, error) {
	jwtParser := &jwt.Parser{}
	claims := make(jwt.MapClaims, 10)
	_, _, err := jwtParser.ParseUnverified(token, &claims)

	return claims, err
}

//HasAuthToken checks if md in ctx has auth
func HasAuthToken(ctx context.Context) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, fmt.Errorf("failed to parse metadata")
	}
	auth := md.Get("grpcgateway-authorization")
	if len(auth) == 0 {
		auth = md.Get("authorization")
		if len(auth) == 0 {
			return true, nil
		}
	}

	return false, nil
}

//UserIDFromContent provides the user ID from the auth found in the given context
func UserIDFromContent(ctx context.Context) (string, error) {
	claims, err := TokenClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}

	return claims["sub"].(string), nil
}

//AccountFromContent provides the current account from the auth found in the given context
func AccountFromContent(ctx context.Context) (string, error) {
	claims, err := TokenClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}

	return claims["acn"].(string), nil
}
