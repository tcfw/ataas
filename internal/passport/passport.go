package passport

import (
	"context"
	"fmt"
	"reflect"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	passportAPI "pm.tcfw.com.au/source/ataas/api/pb/passport"
	"pm.tcfw.com.au/source/ataas/api/pb/users"
	"pm.tcfw.com.au/source/ataas/db"
	"pm.tcfw.com.au/source/ataas/internal/broadcast"
	migrate "pm.tcfw.com.au/source/ataas/internal/passport/db"
	authUtils "pm.tcfw.com.au/source/ataas/internal/passport/utils"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

//NewServer creates a ne struct to interface the auth server
func NewServer(ctx context.Context) (*Server, error) {
	log := logrus.New()
	s := &Server{
		log:     log,
		limiter: &limiter{log: log},
	}

	err := s.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

//Server basic passport construct
type Server struct {
	passportAPI.UnimplementedPassportSeviceServer

	log     *logrus.Logger
	limiter *limiter
}

//Migrate updates the DB schema
func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

//VerifyToken takes in a VerifyTokenRequest, validates the token in that request
func (s *Server) VerifyToken(ctx context.Context, request *passportAPI.VerifyTokenRequest) (*passportAPI.VerifyTokenResponse, error) {
	return s.verifyToken(ctx, request)
}

//Authenticate takes in oneof a authentication types and tries to generate tokens
func (s *Server) Authenticate(ctx context.Context, request *passportAPI.AuthRequest) (*passportAPI.AuthResponse, error) {
	b, err := broadcast.Driver()
	if err != nil {
		return nil, err
	}

	md, _ := metadata.FromIncomingContext(ctx)
	remoteIP := rpcUtils.RemoteIPFromContext(ctx)

	ok, ttl, _ := s.limiter.CheckIP(ctx, remoteIP)
	if !ok {
		return s.limiter.ReachedResp(ctx, remoteIP, ttl)
	}

	var extraClaims map[string]interface{}

	switch authType := request.Creds.(type) {
	case *passportAPI.AuthRequest_UserCreds:
		creds := request.GetUserCreds()
		username := creds.GetUsername()

		//Rate limit
		ok, ttl, remaining := s.limiter.CheckUser(ctx, username, remoteIP)
		if !ok {
			return s.limiter.ReachedResp(ctx, remoteIP, ttl)
		}

		if creds.Recaptcha == "" && creds.MFA == "" {
			return s.limiter.IncreaseResp(ctx, remaining, remoteIP, username, "bad request")
		}

		if creds.Recaptcha != "" {
			valid, err := validateReCAPTCHA(ctx, creds.Recaptcha, remoteIP.String())
			if err != nil {
				return nil, err
			}
			if !valid {
				return s.limiter.IncreaseResp(ctx, remaining, remoteIP, username, "bad request")
			}
		}

		//Find User
		usersSvc, err := usersSvc()
		if err != nil {
			return nil, err
		}

		user, err := usersSvc.Find(withAuthContext(ctx), &users.UserRequest{Query: &users.UserRequest_Email{Email: username}, Status: users.UserRequest_ACTIVE}, grpc.Header(&md))
		if err != nil {
			if serr, ok := status.FromError(err); ok && serr.Code() <= 16 {
				s.log.WithField("status", serr.Code().String()).Error("failed to find user")
				return nil, fmt.Errorf("RPC failed: %s", serr.Code().String())

			}
			return s.limiter.IncreaseResp(ctx, remaining, remoteIP, username, "Unknown user")
		}

		//TODO(tcfw): Check blocked devFP+UID
		if creds.InsecureLogin {
			//Validate password hash
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.GetPassword()))
			if err != nil {
				return s.limiter.IncreaseResp(ctx, remaining, remoteIP, username, "Password mismatch")
			}
		} else {
			//MFA
			if creds.GetMFA() == "" {
				return s.mfaChallenge(ctx, user)
			}

			valid, err := s.validateMFA(ctx, user, request)
			if err != nil {
				return nil, status.Error(500, err.Error())
			}
			if !valid {
				return nil, status.Error(403, "Invalid challenge response")
			}
		}

		// clearRateLimit(username, remoteIP)
		extraClaims = UserClaims(user)

	default:
		b.Publish("passport", &broadcast.AuthenticateEvent{
			Event:    &broadcast.Event{Type: "vanga.passport.authenticate"},
			AuthType: fmt.Sprintf("%v", authType),
			Success:  false,
			Err:      "unknown_type",
		})
		return nil, fmt.Errorf("unknown auth type: %v", authType)
	}

	tokenString, token, err := s.makeNewToken(ctx, extraClaims)
	if err != nil {
		return nil, err
	}
	refreshToken := MakeNewRefresh(ctx)

	b.Publish("passport", &broadcast.AuthenticateEvent{
		Event:    &broadcast.Event{Type: "vanga.passport.authenticate"},
		AuthType: fmt.Sprintf("%v", reflect.TypeOf(request.Creds)),
		Success:  true,
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

//Sessions lists all active tokens under a subject
func (s *Server) Sessions(ctx context.Context, _ *passportAPI.Empty) (*passportAPI.SessionList, error) {
	claims, err := authUtils.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	subject := claims["sub"].(string)

	ret := &passportAPI.SessionList{
		Sessions: []*passportAPI.Session{},
	}

	q := db.Build().Select("jti", "ip", "ua").From("sessions").Where(sq.And{sq.Eq{"sub": subject, "revoked": false}, sq.GtOrEq{"exp": time.Now()}})
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()

	for res.Next() {
		session := &passportAPI.Session{}

		if err := res.Scan(&session.Jti, &session.Ip, &session.UserAgent); err != nil {
			return nil, err
		}

		ret.Sessions = append(ret.Sessions, session)
	}

	return ret, nil
}

func withAuthContext(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	return metadata.NewOutgoingContext(ctx, md)
}
