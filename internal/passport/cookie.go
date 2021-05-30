package passport

import (
	"context"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	sessionCookieName     = "session"
	sessionCookieHTTPOnly = true
	sessionCookieSecure   = true
	authCookieName        = "auth"
	cookiePath            = "/"
)

var (
	cookieDomain = ".ataas.io"
)

func addSessionCookie(ctx context.Context, tokenString string, token *jwt.Token) {
	if d := viper.GetString("passport.cookie_domain"); d != "" {
		cookieDomain = d
	}

	domain := cookieDomain

	// md, ok := metadata.FromIncomingContext(ctx)
	// if ok {
	// 	h := md.Get("grpcgateway-origin")
	// 	if len(h) > 0 {
	// 		url, _ := url.Parse(h[0])
	// 		domain = url.Host
	// 	}
	// }

	sessionCookie := http.Cookie{
		Value: tokenString,

		Domain: domain,
		// Domain:   domain,
		Expires:  time.Unix(token.Claims.(jwt.MapClaims)["exp"].(int64), 0),
		HttpOnly: sessionCookieHTTPOnly,
		Name:     sessionCookieName,
		Path:     cookiePath,
		Secure:   sessionCookieSecure,
		SameSite: http.SameSiteNoneMode,
	}

	authValue := "g2g"

	if _, ok := token.Claims.(jwt.MapClaims)["admin"]; ok {
		authValue = "admin"
	}

	authOKCookie := http.Cookie{
		Value: authValue,

		Domain:   domain,
		Expires:  time.Unix(token.Claims.(jwt.MapClaims)["exp"].(int64), 0),
		HttpOnly: false,
		Secure:   sessionCookieSecure,
		Name:     authCookieName,
		Path:     cookiePath,
		SameSite: http.SameSiteNoneMode,
	}
	grpc.SendHeader(ctx, metadata.Pairs("Set-Cookie", sessionCookie.String(), "Set-Cookie", authOKCookie.String()))
}

func clearSessionCookie(ctx context.Context) {
	if d := viper.GetString("passport.cookie_domain"); d != "" {
		cookieDomain = d
	}

	domain := cookieDomain

	// md, ok := metadata.FromIncomingContext(ctx)
	// if ok {
	// 	h := md.Get("grpcgateway-origin")
	// 	if len(h) > 0 {
	// 		url, _ := url.Parse(h[0])
	// 		domain = url.Host
	// 	}
	// }

	sessionCookie := http.Cookie{
		Value: "gone!",

		Domain:   domain,
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: sessionCookieHTTPOnly,
		Name:     sessionCookieName,
		Path:     cookiePath,
		Secure:   sessionCookieSecure,
		SameSite: http.SameSiteNoneMode,
	}

	authOKCookie := http.Cookie{
		Value: "gone!",

		Domain:   domain,
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: false,
		Secure:   sessionCookieSecure,
		Name:     authCookieName,
		Path:     cookiePath,
		SameSite: http.SameSiteNoneMode,
	}
	grpc.SendHeader(ctx, metadata.Pairs("Set-Cookie", sessionCookie.String(), "Set-Cookie", authOKCookie.String()))
}
