package passport

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	passportAPI "pm.tcfw.com.au/source/trader/api/pb/passport"
	"pm.tcfw.com.au/source/trader/db"
	"pm.tcfw.com.au/source/trader/internal/broadcast"
	authUtils "pm.tcfw.com.au/source/trader/internal/passport/utils"
)

var (
	revoked   map[string]revokedEntry = make(map[string]revokedEntry, 500)
	revokedMu sync.RWMutex
	stopWatch chan struct{} = make(chan struct{})
)

type revokedEntry struct {
	expires time.Time
}

//RevokeToken adds a token to the revoked tokens list
func (s *Server) RevokeToken(ctx context.Context, request *passportAPI.Revoke) (*passportAPI.Empty, error) {
	token, err := authUtils.GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.VerifyToken(ctx, &passportAPI.VerifyTokenRequest{Token: token})
	if err != nil || resp.Revoked || !resp.Valid {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorised")
	}

	claims, err := authUtils.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	request, err = s.populateRevokeRequestDefaults(request, claims)
	if err != nil {
		return nil, err
	}

	err = s.revokeToken(ctx, claims, request.Reason)
	if err != nil {
		return nil, err
	}

	clearSessionCookie(ctx)

	return &passportAPI.Empty{}, nil
}

func (s *Server) populateRevokeRequestDefaults(req *passportAPI.Revoke, claims map[string]interface{}) (*passportAPI.Revoke, error) {
	if req.Id == "" {
		req.Id = claims["sub"].(string)
	}

	if req.Jti == "" {
		req.Jti = claims["jti"].(string)
	}

	if _, ok := claims["admin"]; req.Id != claims["sub"].(string) && !ok {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorised")
	}

	if req.Reason == "" {
		req.Reason = "LOGOUT"
	}

	return req, nil
}

//RevokeAllTokens adds all tokens under a subject to the revoked tokens list
func (s *Server) RevokeAllTokens(ctx context.Context, _ *passportAPI.Empty) (*passportAPI.Empty, error) {
	token, err := authUtils.GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.VerifyToken(ctx, &passportAPI.VerifyTokenRequest{Token: token})
	if err != nil || resp.Revoked || !resp.Valid {
		return nil, status.Errorf(codes.PermissionDenied, "Unauthorised")
	}

	claims, err := authUtils.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.revokeAll(ctx, claims["sub"].(string), "revoke all")
	if err != nil {
		return nil, err
	}

	return &passportAPI.Empty{}, nil
}

func (s *Server) isTokenRevoked(ctx context.Context, claims map[string]interface{}) (bool, error) {
	revokedMu.RLock()
	defer revokedMu.RUnlock()

	if _, ok := revoked[claims["jti"].(string)]; ok {
		return true, nil
	}

	return false, nil
}

func (s *Server) revokeToken(ctx context.Context, claims map[string]interface{}, reason string) error {
	exp := time.Unix(int64(claims["exp"].(float64)), 0)

	if err := s.storeRevoked(ctx, claims["jti"].(string), exp); err != nil {
		return err
	}

	b, err := broadcast.Driver()
	if err != nil {
		return err
	}

	b.Publish("passport", &broadcast.Event{
		Type: "ataas.passport.revoked",
		Metadata: map[string]interface{}{
			"jti":    claims["jti"],
			"sub":    claims["sub"],
			"exp":    claims["exp"],
			"reason": reason,
		},
	})

	return nil
}

func (s *Server) revokeAll(ctx context.Context, subject string, reason string) error {
	q := db.Build().Select("jti", "sub", "exp").From("sessions").Where(sq.And{sq.Eq{"revoked": false, "sub": subject}, sq.GtOrEq{"exp": time.Now()}})
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return err
	}
	defer done()

	var jti string
	var sub string
	var exp time.Time

	for res.Next() {
		if err := res.Scan(&jti, &sub, &exp); err != nil {
			return err
		}
		s.revokeToken(ctx, map[string]interface{}{
			"jti": jti,
			"sub": sub,
			"exp": exp.Unix(),
		}, reason)
	}

	return nil
}

func (s *Server) storeRevoked(ctx context.Context, jti string, expr time.Time) error {
	q := db.Build().Update("sessions").SetMap(sq.Eq{"revoked": true}).Where(sq.Eq{"jti": jti})
	err := db.SimpleExec(ctx, q)
	if err != nil {
		return err
	}

	revoked[jti] = revokedEntry{
		expires: expr,
	}

	return nil
}

func (s *Server) watchRevoke() error {
	ev, close, err := broadcast.ListenForBroadcast("", "vanga.passport.revoked", "")
	if err != nil {
		return err
	}
	defer close()

	gcTick := time.NewTicker(1 * time.Minute)

	for {
		select {
		case eventJSON := <-ev:
			evnt := &broadcast.Event{}
			err := json.Unmarshal(eventJSON, evnt)
			if err != nil {
				s.log.WithError(err).Warn("failed to decode event")
				continue
			}
			jti := evnt.Metadata["jti"].(string)
			revoked[jti] = revokedEntry{
				expires: time.Unix(int64(evnt.Metadata["exp"].(float64)), 0),
			}
			s.log.WithField("jti", jti).Debugf("received peer revoke")
		case <-gcTick.C:
			s.cleanupRevoked()
		case <-stopWatch:
			return nil
		}
	}
}

func (s *Server) repopulateRevoked(ctx context.Context) error {
	q := db.Build().Select("jti", "exp").From("sessions").Where(sq.And{sq.Eq{"revoked": true}, sq.Gt{"exp": time.Now()}})
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return err
	}
	defer done()

	var jti string
	var expr time.Time

	for res.Next() {
		if err := res.Scan(&jti, &expr); err != nil {
			return err
		}

		revoked[jti] = revokedEntry{expires: expr}
	}

	return nil
}

func (s *Server) cleanupRevoked() {
	revokedMu.Lock()
	defer revokedMu.Unlock()

	for k, ex := range revoked {
		if ex.expires.Before(time.Now()) {
			delete(revoked, k)
		}
	}

	q := db.Build().Delete("sessions").Where(sq.Lt{"exp": time.Now()})
	err := db.SimpleExec(context.Background(), q)
	if err != nil {
		s.log.WithError(err).Error("failed to gc sessions table")
	}
}
