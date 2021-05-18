package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	fb "github.com/huandu/facebook/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	passportAPI "pm.tcfw.com.au/source/ataas/api/pb/passport"
	usersAPI "pm.tcfw.com.au/source/ataas/api/pb/users"
	"pm.tcfw.com.au/source/ataas/internal/broadcast"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

type socialUserInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

//SocialLogin validates remote idP tokens and creates users and passes back auth tokens
func (s *Server) SocialLogin(ctx context.Context, request *passportAPI.SocialRequest) (*passportAPI.AuthResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	remoteIP := rpcUtils.RemoteIPFromContext(ctx)

	info, err := getSocialInfo(request)

	if err != nil || info.Email == "" {
		return &passportAPI.AuthResponse{Success: false}, fmt.Errorf("login using provided tokens: %s", err)
	}

	usersSvc, err := usersSvc()
	if err != nil {
		return nil, err
	}

	user, err := usersSvc.Find(withAuthContext(ctx), &usersAPI.UserRequest{Query: &usersAPI.UserRequest_Email{Email: info.Email}}, grpc.Header(&md))
	if err != nil {
		s.limiter.Inc(ctx, info.Email, remoteIP)
		return nil, status.Errorf(codes.Unauthenticated, "incorrect username or passport")
	}

	extraClaims := UserClaims(user)
	s.limiter.Clear(ctx, info.Email, remoteIP)

	tokenString, token, _ := s.makeNewToken(ctx, extraClaims)
	refreshToken := MakeNewRefresh(ctx)

	b, err := broadcast.Driver()
	if err != nil {
		return nil, err
	}
	b.Publish("passport", broadcast.AuthenticateEvent{
		Event:    &broadcast.Event{Type: "vanga.passport.authenticate"},
		AuthType: request.GetProvider(),
		Success:  true,
		User:     user.Id,
		IP:       remoteIP.String(),
	})

	addSessionCookie(ctx, tokenString, token)

	return &passportAPI.AuthResponse{
		Success: true,
		Tokens: &passportAPI.Tokens{
			Token:         "",
			TokenExpire:   token.Claims.(jwt.MapClaims)["exp"].(int64),
			RefreshToken:  *refreshToken,
			RefreshExpire: time.Now().Add(time.Hour * 8).Unix(),
		},
	}, nil
}

func getSocialInfo(request *passportAPI.SocialRequest) (*socialUserInfo, error) {
	switch provider := request.GetProvider(); provider {
	case "google":
		return validateGoogleLogin(request)
	case "facebook":
		return validateFacebookLogin(request)
	default:
		return nil, fmt.Errorf("Unknown social provider %s", provider)
	}
}

func validateFacebookLogin(request *passportAPI.SocialRequest) (*socialUserInfo, error) {
	res, err := fb.Get("/me", fb.Params{
		"fields":       "first_name,last_name,email",
		"access_token": request.GetIdpTokens().GetToken(),
	})
	if err != nil {
		return nil, err
	}

	userInfo := &socialUserInfo{
		Name:  fmt.Sprintf("%s %s", res["first_name"], res["last_name"]),
		Email: res.GetField("email").(string),
	}

	return userInfo, nil
}

func validateGoogleLogin(request *passportAPI.SocialRequest) (*socialUserInfo, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=%s", request.GetIdpTokens().GetToken()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	userInfo := &socialUserInfo{}
	json.NewDecoder(resp.Body).Decode(userInfo)

	return userInfo, nil
}
